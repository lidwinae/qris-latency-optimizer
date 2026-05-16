import http from 'k6/http';
import { check, sleep } from 'k6';

// Menangkap label skenario dari terminal
const scenarioName = __ENV.SCENARIO || 'Default_Test';

export function setup() {
    // Membuat 1 transaksi di awal untuk di-polling oleh ratusan user
    const payload = JSON.stringify({
        qr_payload: "DUMMY_QR_PAYLOAD_123",
        merchant_id: "5626a17b-b3eb-46fc-aab3-50b23f71ec39", 
        amount: 50000
    });
    
    const params = { headers: { 'Content-Type': 'application/json' } };
    const res = http.post('http://localhost:8080/api/transactions/scan', payload, params);
    
    let txId = "";
    if (res.status === 201) {
        txId = res.json().data.transaction_id;
        console.log(`[SETUP] Transaksi ID: ${txId} siap di-polling untuk skenario: ${scenarioName}`);
    }
    return { transactionId: txId };
}

export const options = {
    vus: 300,        // 300 user melakukan cek status bersamaan
    duration: '30s', // Hajar selama 30 detik
    thresholds: {
        http_req_duration: ['p(95)<3000'],
    },
};

export default function (data) {
    if (!data.transactionId) return;

    // Hit endpoint API dan tambahkan TAG (Label) untuk InfluxDB & Grafana
    const res = http.get(`http://localhost:8080/api/transactions/${data.transactionId}`, {
        tags: { my_scenario: scenarioName } 
    });
    
    check(res, {
        'Status is 200 OK': (r) => r.status === 200,
    });

    sleep(1); 
}