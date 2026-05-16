import http from 'k6/http';
import { check, sleep } from 'k6';

// Menangkap label skenario dari terminal
const scenarioName = __ENV.SCENARIO || 'Event_Driven_Async';

export const options = {
    stages: [
        { duration: '10s', target: 50 }, // Naik bertahap ke 50 Virtual Users
        { duration: '30s', target: 50 }, // Tahan beban selama 30 detik
        { duration: '10s', target: 0 },  // Selesai
    ],
    thresholds: {
        // Target Super Ketat! Karena pakai RabbitMQ, harus di bawah 100ms
        http_req_duration: ['p(95)<100'], 
    },
};

export default function () {
    // 1. INQUIRY (Simulasi Scan QR)
    const payload = JSON.stringify({
        qr_payload: "DUMMY_QR_PAYLOAD_123",
        merchant_id: "5626a17b-b3eb-46fc-aab3-50b23f71ec39", // UUID Kantin FILKOM UB
        amount: 15000
    });
    
    const params = { headers: { 'Content-Type': 'application/json' } };
    const scanRes = http.post('http://localhost:8080/api/transactions/scan', payload, params);

    let txId = null;
    if (scanRes.status === 201) {
        txId = scanRes.json().data.transaction_id;
    }

    // Jeda sebentar (simulasi user sedang membaca tagihan & masukkin PIN)
    sleep(1.5); 

    // 2. PAYMENT CONFIRMATION (Simulasi Klik Bayar)
    if (txId) {
        // Tambahkan tag scenarioName agar terpisah di Grafana
        const confirmRes = http.post(`http://localhost:8080/api/transactions/${txId}/confirm`, null, {
            tags: { my_scenario: scenarioName }
        });
        
        check(confirmRes, {
            'Payment Confirmed (200 OK)': (r) => r.status === 200,
            'Status is PROCESSING': (r) => r.json().data.status === "PROCESSING",
        });
    }
    
    sleep(1);
}