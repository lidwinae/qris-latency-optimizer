import { useState } from 'react';
import QRScanner from './components/QRScanner';
import PaymentForm from './components/PaymentForm';
import TransactionStatus from './components/TransactionStatus';
import { scanQR } from './services/api';
import './App.css';

function App() {
  const [screen, setScreen] = useState('scanner'); // scanner | form | status
  const [scannedQR, setScannedQR] = useState('');
  const [transactionId, setTransactionId] = useState('');
  const [loading, setLoading] = useState(false);

  const handleQRScanned = (qrData) => {
    console.log('QR Scanned:', qrData);
    setScannedQR(qrData);
    setScreen('form');
  };

  const handlePaymentSubmit = async (formData) => {
    setLoading(true);
    try {
      const response = await scanQR(
        formData.qrPayload,
        formData.merchantId,
        formData.amount
      );

      if (response.data && response.data.transaction_id) {
        setTransactionId(response.data.transaction_id);
        setScreen('status');
      } else {
        alert('Failed to create transaction');
      }
    } catch (error) {
      console.error('Error:', error);
      alert('Error: ' + (error.response?.data?.error || error.message));
    } finally {
      setLoading(false);
    }
  };

  const handleBack = () => {
    setScannedQR('');
    setTransactionId('');
    setScreen('scanner');
  };

  return (
    <div className="app">
      {screen === 'scanner' && (
        <QRScanner onScan={handleQRScanned} isScanning={true} />
      )}

      {screen === 'form' && (
        <PaymentForm
          onSubmit={handlePaymentSubmit}
          isLoading={loading}
          scannedQR={scannedQR}
        />
      )}

      {screen === 'status' && (
        <TransactionStatus
          transactionId={transactionId}
          onBack={handleBack}
        />
      )}
    </div>
  );
}

export default App;