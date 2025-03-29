 // Format wallet balance on page load
 let amountElement = document.getElementById("amount");
 let rawAmount = amountElement.innerText.replace("₹", "").trim();
 let amount = parseFloat(rawAmount);
 let formattedAmount = new Intl.NumberFormat('en-IN', {
     minimumFractionDigits: 2,
     maximumFractionDigits: 2
 }).format(amount);
 amountElement.innerHTML = `₹${formattedAmount}`;

 // Toggle forms visibility
 document.getElementById('addMoneyBtn').addEventListener('click', function () {
     document.getElementById('addMoneyForm').classList.remove('hidden');
     document.getElementById('giftCardForm').classList.add('hidden');
 });
 
 document.getElementById('sendGiftBtn').addEventListener('click', function () {
     document.getElementById('giftCardForm').classList.remove('hidden');
     document.getElementById('addMoneyForm').classList.add('hidden');
 });

 document.getElementById('cancelAddMoney').addEventListener('click', function () {
     document.getElementById('addMoneyForm').classList.add('hidden');
 });

 document.getElementById('cancelGiftCard').addEventListener('click', function () {
     document.getElementById('giftCardForm').classList.add('hidden');
 });

 // Wallet and Gift Card form logic
 document.addEventListener('DOMContentLoaded', function () {
     // Wallet Form Logic
     const amountButtons = document.querySelectorAll('.amount-btn');
     const otherAmountBtn = document.getElementById('otherAmountBtn');
     const customAmountSection = document.getElementById('customAmountSection');
     const amountInput = document.getElementById('customAmountInput');
     const selectedAmountField = document.getElementById('selectedAmount');
     const submitButton = document.getElementById('submitButton');
     const cancelButton = document.getElementById('cancelAddMoney');
     const statusMessage = document.getElementById('statusMessage');

     // Wallet Helper functions
     function enableSubmitButton() {
         submitButton.disabled = false;
         submitButton.classList.remove('opacity-50');
     }

     function disableSubmitButton() {
         submitButton.disabled = true;
         submitButton.classList.add('opacity-50');
     }

     function resetForm() {
         amountInput.value = '';
         document.getElementById('payment-method').selectedIndex = 0;
         amountButtons.forEach(btn => btn.classList.remove('bg-blue-500', 'text-white'));
         otherAmountBtn.classList.remove('bg-blue-500', 'text-white');
         customAmountSection.classList.add('hidden');
         selectedAmountField.value = '';
         disableSubmitButton();
         submitButton.textContent = 'Add Money';
         statusMessage.classList.add('hidden');
     }

     disableSubmitButton();

     amountButtons.forEach(button => {
         button.addEventListener('click', function () {
             amountButtons.forEach(btn => btn.classList.remove('bg-blue-500', 'text-white'));
             otherAmountBtn.classList.remove('bg-blue-500', 'text-white');
             this.classList.add('bg-blue-500', 'text-white');
             customAmountSection.classList.add('hidden');
             selectedAmountField.value = this.getAttribute('data-amount');
             enableSubmitButton();
             statusMessage.classList.add('hidden');
         });
     });

     otherAmountBtn.addEventListener('click', function () {
         amountButtons.forEach(btn => btn.classList.remove('bg-blue-500', 'text-white'));
         this.classList.add('bg-blue-500', 'text-white');
         customAmountSection.classList.remove('hidden');
         amountInput.value = '';
         selectedAmountField.value = '';
         amountInput.focus();
         disableSubmitButton();
         statusMessage.classList.add('hidden');
     });

     function validateCustomAmount() {
         const value = amountInput.value.trim();
         const amount = parseFloat(value);
         if (value && !isNaN(amount) && amount >= 1) {
             selectedAmountField.value = amount;
             enableSubmitButton();
         } else {
             selectedAmountField.value = '';
             disableSubmitButton();
         }
     }

     if (amountInput) {
         amountInput.addEventListener('input', validateCustomAmount);
         amountInput.addEventListener('change', validateCustomAmount);
     }

     submitButton.addEventListener('click', function () {
         if (!customAmountSection.classList.contains('hidden')) {
             validateCustomAmount();
         }

         const amount = selectedAmountField.value;
         if (!amount || isNaN(parseFloat(amount)) || parseFloat(amount) < 1) {
             showErrorToast('Please select or enter a valid amount');
             return;
         }

         disableSubmitButton();
         submitButton.textContent = 'Processing...';

         const data = {
             amount: parseFloat(amount),
             paymentMethod: document.getElementById('payment-method').value
         };

         fetch('/profile/wallet/add/amount', {
             method: 'POST',
             headers: {
                 'Content-Type': 'application/json',
                 'X-CSRF-TOKEN': document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || ''
             },
             body: JSON.stringify(data)
         })
             .then(response => {
                 if (!response.ok) {
                     throw new Error('Network response was not ok');
                 }
                 return response.json();
             })
             .then(data => {
                 showSuccessToast("Payment initiated successfully!");
                 initializeRazorpay(data);
                 setTimeout(() => {
                     resetForm();
                     document.getElementById('addMoneyForm').classList.add('hidden');
                 }, 3000);
             })
             .catch(error => {
                 console.error("Payment initiation failed:", error);
                 showErrorToast('Failed to add money. Please try again.');
                 enableSubmitButton();
                 submitButton.textContent = 'Add Money';
             });
     });

     cancelButton.addEventListener('click', resetForm);

     // Gift Card Form Logic
     const giftCardForm = document.getElementById('giftCardForm');
     const sendGiftButton = giftCardForm.querySelector('button[type="submit"]');
     const recipientName = document.getElementById('recipient-name');
     const recipientEmail = document.getElementById('recipient-email');
     const giftAmount = document.getElementById('gift-amount');
     const giftMessage = document.getElementById('gift-message');

     function enableGiftButton() {
         sendGiftButton.disabled = false;
         sendGiftButton.classList.remove('opacity-50');
     }

     function disableGiftButton() {
         sendGiftButton.disabled = true;
         sendGiftButton.classList.add('opacity-50');
     }

     function resetGiftForm() {
         recipientName.value = '';
         recipientEmail.value = '';
         giftAmount.value = '';
         giftMessage.value = '';
         enableGiftButton();
         sendGiftButton.textContent = 'Send Gift Card';
     }

     enableGiftButton();

     giftCardForm.addEventListener('submit', function (e) {
         e.preventDefault(); // Prevent default form submission

         // Client-side validation
         if (!recipientName.value.trim()) {
             showErrorToast('Please enter recipient name');
             return;
         }
         if (!recipientEmail.value.trim() || !/^\S+@\S+\.\S+$/.test(recipientEmail.value)) {
             showErrorToast('Please enter a valid email');
             return;
         }
         const amount = parseFloat(giftAmount.value);
         if (!amount || amount < 5) {
             showErrorToast('Please enter an amount of at least ₹5');
             return;
         }

         disableGiftButton();
         sendGiftButton.textContent = 'Sending...';

         const giftData = {
             recipient_name: recipientName.value.trim(),
             recipient_email: recipientEmail.value.trim(),
             amount: amount,
             message: giftMessage.value.trim()
         };

         fetch('/profile/wallet/send/gift/card', {
             method: 'POST',
             headers: {
                 'Content-Type': 'application/json',
                 'X-CSRF-TOKEN': document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || '',
                 'Accept': 'application/json'
             },
             body: JSON.stringify(giftData)
         })
             .then(response => {
                 // Get the JSON data regardless of success or failure
                 return response.json().then(data => ({
                     data,
                     status: response.status,
                     ok: response.ok
                 }));
             })
             .then(({ data, ok }) => {
                 if (!ok) {
                     // Use error message from backend if available, otherwise use default
                     const errorMessage = data.message || data.error || 'Failed to send gift card';
                     throw new Error(errorMessage);
                 }

                 // Use success message from backend if available, otherwise use default
                 const successMessage = data.message || 'Gift card sent successfully!';
                 showSuccessToast(successMessage);
                 resetGiftForm();
                 setTimeout(() => {
                     document.getElementById('giftCardForm').classList.add('hidden');
                     window.location.reload();
                 }, 3000);
             })
             .catch(error => {
                 console.error('Gift card sending failed:', error);
                 // Display the error message from the backend
                 showErrorToast(error.message || 'Failed to send gift card. Please try again.');
                 enableGiftButton();
                 sendGiftButton.textContent = 'Send Gift Card';
             });
     });
 });

 // Razorpay integration
 function initializeRazorpay(data) {
     const options = {
         key: data.key_id,
         amount: data.amount,
         currency: data.currency,
         name: "LAPTIX",
         description: "Purchase Payment",
         image: "https://res.cloudinary.com/dghzlcoco/image/upload/v1740498507/text-1740498489427_ir9mat.png",
         order_id: data.order_id,
         handler: function (response) {
             verifyPayment({
                 razorpay_payment_id: response.razorpay_payment_id,
                 razorpay_order_id: response.razorpay_order_id,
                 razorpay_signature: response.razorpay_signature,
                 amount: data.amount,
             });
         },
         prefill: {
             name: data.prefill.name,
             email: data.prefill.email,
         },
         notes: data.notes,
         theme: {
             color: "#0000"
         },
         modal: {
             ondismiss: function () {
                 showErrorToast("Payment cancelled");
                 document.getElementById('submitButton').textContent = 'Add Money';
                 enableSubmitButton();
             }
         },
         retry: {
             enabled: false
         }
     };

     const rzp = new Razorpay(options);
     rzp.on('payment.failed', function (response) {
         showErrorToast("Payment Failed");
     });
     rzp.open();
 }

 function verifyPayment(paymentData) {
     fetch('/profile/wallet/add/amount/verify', {
         method: 'POST',
         headers: {
             'Content-Type': 'application/json',
             'X-CSRF-TOKEN': document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || ''
         },
         body: JSON.stringify(paymentData),
     })
         .then(response => {
             if (!response.ok) {
                 return response.json().then(err => { throw new Error(err.message || "Payment verification failed."); });
             }
             return response.json();
         })
         .then(data => {
             showSuccessToast("Amount Added Successfully");
             setTimeout(() => {
                 window.location.reload();
             }, 3000);
         })
         .catch(error => {
             console.error('Verification error:', error);
             showErrorToast(error.message || "An error occurred while verifying the payment.");
             document.getElementById('submitButton').textContent = 'Add Money';
             enableSubmitButton();
         });
 }

 // Toast functions
 function showSuccessToast(message) {
     const toast = document.getElementById('toast');
     const toastMessage = document.querySelector('.toast-message');
     const successIcon = document.querySelector('.toast-icon-success');
     const errorIcon = document.querySelector('.toast-icon-error');
     toastMessage.textContent = message;
     successIcon.classList.remove('hidden');
     errorIcon.classList.add('hidden');
     toast.classList.add('show');
     setTimeout(() => {
         toast.classList.remove('show');
         successIcon.classList.add('hidden');
     }, 3000);
 }

 function showErrorToast(message) {
     const toast = document.getElementById('toast');
     const toastMessage = document.querySelector('.toast-message');
     const successIcon = document.querySelector('.toast-icon-success');
     const errorIcon = document.querySelector('.toast-icon-error');
     toastMessage.textContent = message;
     successIcon.classList.add('hidden');
     errorIcon.classList.remove('hidden');
     toast.classList.add('show');
     setTimeout(() => {
         toast.classList.remove('show');
         errorIcon.classList.add('hidden');
     }, 3000);
 }