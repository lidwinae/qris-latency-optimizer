import http from 'k6/http';
import { check, sleep } from 'k6';

// Otomatis memberi nama skenario "Tanpa_Cache" untuk kotak merah di Grafana
const scenarioName = __ENV.SCENARIO || 'Tanpa_Cache';

export const options = {
    stages: [
        { duration: '10s', target: 50 },
        { duration: '30s', target: 50 },
        { duration: '10s', target: 0 }, 
    ],
    thresholds: {
        http_req_duration: ['p(95)<3000'], 
    },
};

export default function () {
    // 1. INQUIRY 
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

    sleep(1.5); 

    // 2. PAYMENT CONFIRMATION (Menembak Endpoint yang Lambat: /confirm-sync)
    if (txId) {
        // Perhatikan URL-nya berakhiran /confirm-sync
        const confirmRes = http.post(`http://localhost:8080/api/transactions/${txId}/confirm-sync`, null, {
            tags: { my_scenario: scenarioName } // Mengirim tag "Tanpa_Cache"
        });
        
        check(confirmRes, {
            'Payment Confirmed (200 OK)': (r) => r.status === 200,
            'Status is SUCCESS': (r) => r.json().data.status === "SUCCESS",
        });
    }
    
    sleep(1);
}