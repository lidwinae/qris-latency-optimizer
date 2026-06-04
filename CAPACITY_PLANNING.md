# Laporan Perencanaan Kapasitas (Capacity Planning): QRIS Latency Optimizer

Dokumen ini menguraikan strategi perencanaan kapasitas untuk sistem QRIS Latency Optimizer. Laporan ini didasarkan pada data empiris yang dikumpulkan selama pengujian beban (*load testing*) menggunakan K6 dan Go CLI, serta pemantauan perangkat keras melalui Prometheus dan Node Exporter.

## 1. Baseline Saat Ini & Metrik Performa
Berdasarkan *benchmarking* lokal menggunakan 11-container Docker Compose pada satu mesin tunggal, sistem mencapai baseline berikut:

- **Throughput Maksimal (Lokal):** ~1.300 - 1.600 Requests Per Second (RPS)
- **Tingkat Error (Error Rate):** 0% (Tidak ada data yang hilang di bawah beban berat)
- **P95 Latency:** < 90ms untuk pemrosesan asinkron dari ujung ke ujung (end-to-end)
- **Desain Arsitektur:** `Synchronous API` -> `RabbitMQ Queue` -> `Asynchronous Worker` -> `PostgreSQL / Redis` -> `WebSocket Notification`

*Catatan: 1.300 RPS setara dengan sekitar 112 juta transaksi per hari jika dipertahankan konstan, yang mana sudah merupakan baseline performa yang sangat tinggi untuk sebuah node tunggal.*

---

## 2. Analisis Bottleneck (Batas Sistem)
Jika lalu lintas berskala secara signifikan (misalnya, lalu lintas tingkat nasional sebesar 5.000+ RPS), arsitektur *single-node* saat ini akan mengalami kemacetan (*bottleneck*) dalam urutan berikut:

### A. Layer Komputasi (Golang API & WebSocket)
- **Batas:** Terikat pada CPU (CPU bound). Golang sangat efisien, tetapi menangani ribuan koneksi WebSocket secara bersamaan dan proses konversi JSON (marshalling/unmarshalling) pada akhirnya akan memicu lonjakan penggunaan CPU hingga 100%.
- **Gejala:** P95 Latency akan mulai menurun dari <100ms menjadi >1.000ms sebelum sistem mengalami *crash*.

### B. Layer Data (PostgreSQL)
- **Batas:** Terikat pada I/O dan Koneksi. Database relasional adalah komponen yang paling sulit untuk di-skalakan secara horizontal untuk proses penulisan (*writes*). Batas `max_connections` bawaan (biasanya 100) akan cepat habis.
- **Gejala:** Muncul error "Too many clients already"; antrean database mengunci (locks) yang mengakibatkan kegagalan transaksi.

### C. Message Broker (RabbitMQ)
- **Batas:** Terikat pada Memori (Memory bound). RabbitMQ bertindak sebagai peredam kejut (*shock absorber*). Jika Go Worker yang memproses penulisan ke database lebih lambat daripada request API yang masuk, ukuran antrean akan terus bertambah.
- **Gejala:** Konsumsi RAM tinggi. Jika RAM habis, RabbitMQ akan memicu alarm memori dan memblokir *publisher* (API akan tertahan/hang).

---

## 3. Strategi Skala Bertingkat (Kesiapan Produksi)

Untuk mempersiapkan lingkungan Produksi dan lalu lintas tinggi, arsitektur harus bertransisi dari pengaturan monolitik Docker-Compose ke lingkungan yang terdistribusi dan skalabel (misalnya, Kubernetes).

### Tingkat 1: Layer Aplikasi / Komputasi
*Tujuan: Menangani permintaan API bersamaan dan WebSockets dalam jumlah masif.*
- **Tindakan:** Deploy Golang Backend sebagai *stateless container* melalui Kubernetes (K8s).
- **Autoscaling (HPA):** Konfigurasikan Horizontal Pod Autoscaler (HPA) untuk memutar pod Go API baru saat rata-rata `process_cpu_seconds_total` (dipantau via Prometheus) melebihi **70%**.
- **Pemisahan Tugas (Separation of Concerns):** Deploy HTTP API Handler dan RabbitMQ Worker di **pod yang terpisah**. Jika antrean menjadi terlalu panjang, Anda hanya perlu meningkatkan skala pod *Worker* untuk memproses penulisan database lebih cepat, sehingga menghemat sumber daya komputasi API.

### Tingkat 2: Layer Database (PostgreSQL)
*Tujuan: Mencegah kehabisan Koneksi DB dan I/O locks.*
- **Connection Pooling:** Implementasikan **PgBouncer** di antara Go Workers dan PostgreSQL. Ini akan melipatgandakan ribuan koneksi aplikasi yang masuk menjadi hanya beberapa lusin koneksi database aktual.
- **Master-Slave Replication:** Arahkan semua operasi `WRITE` (membuat transaksi) ke Primary Master DB. Arahkan semua operasi `READ` (mengambil data merchant, memeriksa status) ke instance *Read-Replica* untuk melepaskan beban CPU/Disk I/O dari master.

### Tingkat 3: Message Queue & Caching (RabbitMQ & Redis)
*Tujuan: Memastikan tidak ada data yang hilang (zero data loss) selama lonjakan lalu lintas (misal: Flash Sale).*
- **RabbitMQ Clustering:** Deploy RabbitMQ dalam kluster ketersediaan tinggi (minimal 3 node) dengan *Quorum Queues*. Hal ini memastikan bahwa jika satu server mati, transaksi yang diantrekan akan tetap aman di node lain.
- **Redis Cluster:** Gunakan Redis untuk *distributed caching*. Karena validasi QRIS sangat bergantung pada pencarian data merchant, tingkat *cache hit* >95% akan menyelamatkan Postgres DB dari beban baca yang masif. Pastikan Redis dikonfigurasi dengan batas memori yang memadai dan kebijakan eviksi (`volatile-lru`).

---

## 4. Rekomendasi Infrastruktur Cloud
Jika menerapkan sistem ini ke penyedia layanan cloud seperti AWS, berikut adalah pemetaan kapasitas yang direkomendasikan untuk target **5.000 RPS**:

| Komponen | Padanan AWS | Spesifikasi Rekomendasi (Titik Awal) |
| :--- | :--- | :--- |
| **Load Balancer** | AWS ALB (Application Load Balancer) | Managed |
| **Komputasi (Go API & Worker)** | Amazon EKS (Kubernetes) atau ECS | 4-6 Pods / Tasks (2 vCPU, 4GB RAM masing-masing) |
| **Message Broker** | Amazon MQ for RabbitMQ | `mq.m5.large` (Clustered) |
| **Database** | Amazon RDS for PostgreSQL | `db.m6g.xlarge` (4 vCPU, 16GB RAM) + 1 Read Replica |
| **Cache** | Amazon ElastiCache (Redis) | `cache.t4g.medium` (Clustered) |
| **Monitoring** | Amazon Managed Grafana & Prometheus | Managed |

## Kesimpulan
Arsitektur sistem saat ini (yang menggunakan **Pemrosesan Asinkron melalui RabbitMQ** dan **Cache-aside dengan Redis**) sudah sangat kuat secara arsitektural untuk lalu lintas skala perusahaan (*enterprise-scale*). Dengan memindahkan sistem ke platform orkestrasi *container* yang terdistribusi dan memisahkan layer API dari layer Worker, sistem *QRIS Latency Optimizer* dapat di-skalakan secara dinamis untuk memenuhi permintaan lalu lintas yang ekstrim sekalipun, sembari mempertahankan latensi di bawah 100ms.
