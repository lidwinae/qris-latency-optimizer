import http from 'k6/http';
import { check, sleep } from 'k6';
import { textSummary } from 'https://jslib.k6.io/k6-summary/0.1.0/index.js';

const scenarioName = __ENV.SCENARIO || 'Event_Driven_Async';
const DASHBOARD_URL = 'http://localhost:8080/api/monitor/k6';
const MERCHANT_ID = '5626a17b-b3eb-46fc-aab3-50b23f71ec39';
const jsonHeaders = { headers: { 'Content-Type': 'application/json' } };

export const options = {
    stages: [
        { duration: '10s', target: 50 },
        { duration: '30s', target: 50 },
        { duration: '10s', target: 0 },
    ],
    thresholds: {
        http_req_duration: ['p(95)<500'],
    },
};

// Generate a valid QRIS payload once before the test
export function setup() {
    const res = http.get(`http://localhost:8080/api/qris?merchant_id=${MERCHANT_ID}&amount=15000`);
    const qrPayload = res.json().qris_payload;
    console.log(`[SETUP] Valid QRIS payload: ${qrPayload}`);
    return { qrPayload };
}

export default function (data) {
    // 1. Create transaction with valid QR payload
    const scanRes = http.post('http://localhost:8080/api/transactions/scan', JSON.stringify({
        qr_payload: data.qrPayload,
        merchant_id: "TEST001",
        amount: 15000
    }), jsonHeaders);

    let txId = null;
    if (scanRes.status === 201) {
        txId = scanRes.json().data.transaction_id;
    }

    sleep(1.5);

    // 2. Confirm payment via OPTIMIZED endpoint (RabbitMQ async)
    if (txId) {
        const confirmRes = http.post(
            `http://localhost:8080/api/transactions/${txId}/confirm`,
            null,
            { tags: { my_scenario: scenarioName } }
        );

        // Report to dashboard
        http.post(`${DASHBOARD_URL}/data`, JSON.stringify({
            points: [{
                scenario: scenarioName,
                metric: "http_req_duration",
                value: confirmRes.timings.duration,
                status: confirmRes.status,
                error: confirmRes.status !== 200
            }, {
                scenario: scenarioName,
                metric: "vus",
                value: __VU
            }]
        }), jsonHeaders);

        check(confirmRes, {
            'Payment Confirmed (200 OK)': (r) => r.status === 200,
            'Status is PROCESSING': (r) => r.json().data.status === "PROCESSING",
        });
    }

    sleep(1);
}

export function handleSummary(data) {
    const m = data.metrics;
    const dur = m.http_req_duration;
    const reqs = m.http_reqs;
    const chk = m.checks;
    const summary = {
        scenario: scenarioName,
        total_reqs: reqs ? reqs.values.count : 0,
        avg_duration: dur ? dur.values.avg : 0,
        p95_duration: dur ? dur.values['p(95)'] : 0,
        p99_duration: dur ? dur.values['p(99)'] : 0,
        min_duration: dur ? dur.values.min : 0,
        max_duration: dur ? dur.values.max : 0,
        error_rate: m.http_req_failed ? m.http_req_failed.values.rate * 100 : 0,
        throughput: reqs ? reqs.values.rate : 0,
        checks_pass: chk ? chk.values.passes : 0,
        checks_fail: chk ? chk.values.fails : 0,
        data_sent: m.data_sent ? m.data_sent.values.count : 0,
        data_received: m.data_received ? m.data_received.values.count : 0,
        max_vus: m.vus_max ? m.vus_max.values.max : 0,
        duration: 50,
    };
    http.post(`${DASHBOARD_URL}/summary`, JSON.stringify(summary), jsonHeaders);
    return { stdout: textSummary(data, { indent: ' ', enableColors: true }) };
}
