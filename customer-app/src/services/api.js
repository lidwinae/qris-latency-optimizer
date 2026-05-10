import axios from 'axios';

const API_BASE_URL = 'http://localhost:8080/api';

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Scan QR - Create Transaction
export const scanQR = async (qrPayload, merchantId, amount) => {
  try {
    const response = await api.post('/transactions/scan', {
      qr_payload: qrPayload,
      merchant_id: merchantId,
      amount: parseFloat(amount),
    });
    return response.data;
  } catch (error) {
    console.error('Error scanning QR:', error);
    throw error;
  }
};

// Get Transaction Status
export const getTransactionStatus = async (transactionId) => {
  try {
    const response = await api.get(`/transactions/${transactionId}`);
    return response.data;
  } catch (error) {
    console.error('Error getting transaction status:', error);
    throw error;
  }
};

// Confirm Payment
export const confirmPayment = async (transactionId) => {
  try {
    const response = await api.post(`/transactions/${transactionId}/confirm`);
    return response.data;
  } catch (error) {
    console.error('Error confirming payment:', error);
    throw error;
  }
};

export default api;