function moveFocus(current, nextId) {
  if (current.value.length === current.maxLength && nextId) {
    document.getElementById(nextId).focus();
  }
}

function handleSubmit(event) {
  event.preventDefault();
  
  const otp = [1,2,3,4,5,6].map(num => 
    document.getElementById(`otp${num}`).value
  ).join('');
  
  document.getElementById('combinedOtp').value = otp;
  
  event.target.submit();
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

function resendOTP() {
  timeLeft = 30;
  resendButton.classList.add('hidden');
  timerElement.classList.remove('hidden');
  updateTimer();
}

updateTimer();