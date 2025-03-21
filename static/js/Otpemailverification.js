// OTP Management
function moveFocus(current, nextId) {
  if (current.value.length === current.maxLength && nextId) {
    document.getElementById(nextId).focus();
  }
}

async function handleSubmit(event) {
  event.preventDefault();

  const otp = [1, 2, 3, 4, 5, 6].map(num =>
    document.getElementById(`otp${num}`).value
  ).join('');

  document.getElementById('combinedOtp').value = otp;

  try {
    const formData = new FormData(event.target);
    const response = await fetch(event.target.action, {
      method: 'POST',
      body: formData,
      headers: {
        'Accept': 'application/json'
      }
    });

    const data = await response.json();

    if (response.ok) {
      showSuccessToast(data.message||'OTP verified successfully!');
      // Delay redirect to show success message
      setTimeout(() => {
        window.location.href = '/auth/login';
      }, 1500);
    } else {
      // Handle different error cases
      switch (data.error) {
        case 'INVALID_OTP':
          showErrorToast(data.message||'Invalid OTP. Please try again.');
          // Clear OTP fields
          [1, 2, 3, 4, 5, 6].forEach(num => {
            document.getElementById(`otp${num}`).value = '';
          });
          document.getElementById('otp1').focus();
          break;
        case 'OTP_EXPIRED':
          showErrorToast(data.message||'OTP has expired. Please request a new one.');
          // Show resend button immediately
          timeLeft = 0;
          timerElement.classList.add('hidden');
          resendButton.classList.remove('hidden');
          break;
        default:
          showErrorToast(data.message || 'Verification failed. Please try again.');
      }
    }
  } catch (error) {
    showErrorToast(data.message||'Network error. Please check your connection.');
  }

  return false;
}

let timeLeft = 30;
const timerElement = document.getElementById('timer');
const resendButton = document.getElementById('resendButton');

function updateTimer() {
  if (timeLeft > 0) {
    timerElement.textContent = `Resend OTP in ${timeLeft} seconds`;
    timeLeft--;
    setTimeout(updateTimer, 1000);
  } else {
    timerElement.classList.add('hidden');
    resendButton.classList.remove('hidden');
  }
}

async function resendOTP() {
  try {
    const response = await fetch('/auth/signup/otp/resend', {
      method: 'POST',
      headers: {
        'Accept': 'application/json'
      }
    });

    const data = await response.json();

    if (response.ok) {
      showSuccessToast(data.message||'OTP resent successfully!');
      timeLeft = 30;
      resendButton.classList.add('hidden');
      timerElement.classList.remove('hidden');
      updateTimer();
      // Clear existing OTP fields
      [1, 2, 3, 4, 5, 6].forEach(num => {
        document.getElementById(`otp${num}`).value = '';
      });
      document.getElementById('otp1').focus();
    } else {
      showErrorToast(data.message || 'Failed to resend OTP. Please try again.');
    }
  } catch (error) {
    showErrorToast(data.message||'Network error. Please check your connection.');
  }
}

// Initialize timer on page load
document.addEventListener('DOMContentLoaded', () => {
  updateTimer();
});