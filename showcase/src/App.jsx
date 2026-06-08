const features = [
  {
    title: 'Dynamic QRIS Generation',
    desc: 'Merchant dapat memilih toko dan nominal, lalu sistem membuat payload QRIS dinamis berbasis data PostgreSQL.',
  },
  {
    title: 'Redis Caching Layer',
    desc: 'Redis digunakan untuk mempercepat lookup merchant dan polling status transaksi dengan pola cache-aside.',
  },
  {
    title: 'Async Payment Processing',
    desc: 'RabbitMQ memproses konfirmasi pembayaran secara asynchronous sehingga response utama tidak terblokir.',
  },
  {
    title: 'Real-Time Merchant Notification',
    desc: 'Merchant menerima notifikasi pembayaran melalui WebSocket setelah transaksi berhasil diproses.',
  },
  {
    title: 'Performance Monitoring',
    desc: 'Prometheus dan Grafana digunakan untuk memantau latency, runtime metrics, dan resource usage.',
  },
  {
    title: 'Rural Network Simulation',
    desc: 'Toxiproxy mensimulasikan jaringan rural dengan latency, jitter, dan bandwidth limit untuk pengujian performa.',
  },
]

const stack = [
  'Go',
  'Gin',
  'PostgreSQL',
  'Redis',
  'RabbitMQ',
  'React',
  'Vite',
  'Prometheus',
  'Grafana',
  'Toxiproxy',
  'k6',
  'Docker Compose',
]

const team = [
  {
    name: 'Khaif Alfansyuri',
    role: 'Network & API Optimization',
    image: '',
  },
  {
    name: 'Muhammad Bhre Aji Dimas Wisanggeni',
    role: 'Load Testing & System Monitoring',
    image: '',
  },
  {
    name: 'Lidwina Eleonora Dora',
    role: 'Caching Layer & Fullstack Integration',
    image: `${import.meta.env.BASE_URL}team/lidwina.png`,
  },
  {
    name: 'Raihani Syuja Natalie Afifah',
    role: 'Frontend Development',
    image: '',
  },
  {
    name: 'Farrel Firmanditto Azka',
    role: 'Caching Prefetching',
    image: '',
  },
]

const screenshots = [
  {
    title: 'Merchant Dashboard',
    desc: 'Tampilan merchant untuk memilih toko, memasukkan nominal, dan generate QRIS dinamis.',
    image: `${import.meta.env.BASE_URL}screenshots/tes.png`,
  },
  {
    title: 'Customer App',
    desc: 'Tampilan customer untuk scan QRIS, validasi payload, dan konfirmasi pembayaran.',
  },
  {
    title: 'Grafana Monitoring',
    desc: 'Dashboard monitoring untuk latency, CPU, memory, Go runtime, dan request metrics.',
  },
  {
    title: 'Load Test Result',
    desc: 'Bukti pengujian performa menggunakan k6 dan Go Load Test CLI.',
  },
]

function App() {
  return (
    <main>
      <section className="hero">
        <div className="heroText">
          <p className="eyebrow">Capstone Project - Kelompok 4</p>
          <h1>QRIS Latency Optimizer</h1>
          <p className="lead">
            Prototype sistem transaksi QRIS real-time berbasis PostgreSQL,
            Redis caching, RabbitMQ, WebSocket, dan monitoring performa untuk
            menurunkan latency pada alur pembayaran digital.
          </p>
          <div className="actions">
            <a href="https://github.com/lidwinae/qris-latency-optimizer" target="_blank">
              Repository
            </a>
            <a href="#features" className="secondary">
              Lihat Fitur
            </a>
          </div>
        </div>

        <div className="heroPanel">
          <div>
            <span>p95 Latency</span>
            <strong>70ms</strong>
          </div>
          <div>
            <span>Throughput</span>
            <strong>1.617 RPS</strong>
          </div>
          <div>
            <span>Latency Target</span>
            <strong>&lt;100ms</strong>
          </div>
          <div>
            <span>Rural Proxy Delay</span>
            <strong>+500ms</strong>
          </div>
        </div>
      </section>

      <section className="section">
        <div className="sectionHeader">
          <p className="eyebrow">Problem</p>
          <h2>Masalah yang Diselesaikan</h2>
        </div>
        <p className="bodyText">
          Sistem transaksi QRIS real-time dapat mengalami latency tinggi,
          timeout, retry pembayaran, dan keterlambatan notifikasi merchant.
          Permasalahan ini makin terasa ketika sistem harus berhadapan dengan
          integrasi legacy, akses database berulang, dan kondisi jaringan
          terbatas seperti wilayah rural.
        </p>
      </section>

      <section className="section split">
        <div>
          <p className="eyebrow">Solution</p>
          <h2>Pendekatan Solusi</h2>
        </div>
        <div className="solutionList">
          <article>
            <h3>Cache-Aside dengan Redis</h3>
            <p>
              Data merchant dan status transaksi yang sering diakses disimpan
              di Redis untuk mengurangi query berulang ke PostgreSQL.
            </p>
          </article>
          <article>
            <h3>Event-Driven Processing</h3>
            <p>
              RabbitMQ memisahkan proses konfirmasi pembayaran dan notifikasi
              dari jalur HTTP utama agar response lebih cepat.
            </p>
          </article>
          <article>
            <h3>Monitoring dan Load Testing</h3>
            <p>
              Prometheus, Grafana, k6, dan Go Load Test CLI digunakan untuk
              mengevaluasi latency, throughput, dan stabilitas sistem.
            </p>
          </article>
        </div>
      </section>

      <section className="section" id="features">
        <div className="sectionHeader">
          <p className="eyebrow">Features</p>
          <h2>Fitur Utama</h2>
        </div>
        <div className="featureGrid">
          {features.map((feature) => (
            <article className="featureCard" key={feature.title}>
              <h3>{feature.title}</h3>
              <p>{feature.desc}</p>
            </article>
          ))}
        </div>
      </section>

      <section className="section">
        <div className="sectionHeader">
          <p className="eyebrow">Architecture</p>
          <h2>Arsitektur Sistem</h2>
        </div>
        <div className="architecture">
          <div>Merchant Dashboard</div>
          <div>Customer App</div>
          <div>Go Backend API</div>
          <div>PostgreSQL</div>
          <div>Redis</div>
          <div>RabbitMQ</div>
          <div>WebSocket</div>
          <div>Prometheus + Grafana</div>
          <div>Toxiproxy</div>
        </div>
      </section>

      <section className="section">
        <div className="sectionHeader">
          <p className="eyebrow">Stack</p>
          <h2>Teknologi yang Digunakan</h2>
        </div>
        <div className="stackList">
          {stack.map((item) => (
            <span key={item}>{item}</span>
          ))}
        </div>
      </section>

      <section className="section">
        <div className="sectionHeader">
          <p className="eyebrow">Evidence</p>
          <h2>Dokumentasi Demo</h2>
        </div>
        <div className="placeholderGrid">
          {screenshots.map((item) => (
            <article className="placeholder" key={item.title}>
              {item.image ? (
                <img
                  src={item.image}
                  alt={item.title}
                  style={{
                    width: '100%',
                    height: '100%',
                    minHeight: '220px',
                    objectFit: 'cover',
                    borderRadius: '6px',
                  }}
                />
              ) : (
                <div>
                  <h3>{item.title}</h3>
                  <p style={{ maxWidth: '32rem', margin: 0 }}>{item.desc}</p>
                </div>
              )}
            </article>
          ))}
        </div>
      </section>

      <section className="section">
        <div className="sectionHeader">
          <p className="eyebrow">Team</p>
          <h2>Anggota Kelompok</h2>
        </div>
        <div className="teamGrid">
          {team.map((member) => (
            <article className="teamCard" key={member.name}>
              {member.image ? (
                <img src={member.image} alt={member.name} />
              ) : (
                <div className="avatarPlaceholder">
                  {member.name
                    .split(' ')
                    .slice(0, 2)
                    .map((word) => word[0])
                    .join('')}
                </div>
              )}
              <h3>{member.name}</h3>
              <p>{member.role}</p>
            </article>
          ))}
        </div>
      </section>

      <footer>
        <strong>QRIS Latency Optimizer</strong>
        <span>Capstone Project - Fakultas Ilmu Komputer - 2026</span>
      </footer>
    </main>
  )
}

export default App
