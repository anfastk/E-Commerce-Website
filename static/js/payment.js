function updateDeliveryDate() {
    let now = new Date();
    let orderTime = new Date(now);
    let deliveryDate = new Date(orderTime);

    // 3 days add cheyyuka
    deliveryDate.setDate(orderTime.getDate() + 7);

    // Midnight check (12:00 AM aayal next day update cheyyuka)
    if (now.getHours() === 23 && now.getMinutes() === 59 && now.getSeconds() === 59) {
        deliveryDate.setDate(deliveryDate.getDate() + 1);
    }

    // Date format set cheyyuka
    const options = { day: "2-digit", month: "short", year: "numeric" };
    let formattedDeliveryDate = deliveryDate.toLocaleDateString("en-US", options);

    // Remaining time calculate cheyyuka
    let midnight = new Date(now);
    midnight.setHours(23, 59, 59, 999);
    let timeLeft = midnight - now;
    let hoursLeft = Math.floor(timeLeft / (1000 * 60 * 60));
    let minutesLeft = Math.floor((timeLeft % (1000 * 60 * 60)) / (1000 * 60));
    let secondsLeft = Math.floor((timeLeft % (1000 * 60)) / 1000);

    // HTML update cheyyuka
    document.getElementById("delivery-date").innerText = `Arriving ${formattedDeliveryDate}`;
    document.getElementById("order-time-left").innerText = `If you order in the next ${hoursLeft} hours, ${minutesLeft} minutes, and ${secondsLeft} seconds`;
}

// Function call cheyyuka
updateDeliveryDate();

// Live update every second
setInterval(updateDeliveryDate, 1000);


// Get necessary elements and data from HTML
document.addEventListener('DOMContentLoaded', () => {
    const paymentOptions = document.querySelectorAll('.payment-option');
    const walletInput = document.querySelector('.wallet-input');
    const walletPaymentOption = document.getElementById('walletPayment');
    const codPaymentOption = document.getElementById('codPayment');
    const walletBalanceDisplay = walletPaymentOption.querySelector('.payment-title');
    const walletMessageDisplay = walletPaymentOption.querySelector('.flex.items-center.gap-1');
    const giftCardInput = document.querySelector('.wallet-input input');
    const applyButton = document.querySelector('.apply-btn');

    // Get data from hidden inputs
    const total = parseFloat(document.getElementById('total').value) || 0;
    const isCodAvailable = document.getElementById('isCodAvailable').value === 'true';

    // Ensure wallet input is visible initially if needed
    if (walletInput) {
        // Make sure wallet input is properly initialized
        walletInput.style.display = 'block';
    }

    // Fetch wallet balance initially
    fetchWalletBalance();

    // Function to fetch wallet balance via AJAX
    function fetchWalletBalance() {
        fetch('/checkout/check/wallet/balance', {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json'
            }
        })
            .then(response => {
                if (!response.ok) {
                    throw new Error('Failed to fetch wallet balance');
                }
                return response.json();
            })
            .then(data => {
                updateWalletDisplay(data.balance);
            })
            .catch(error => {
                console.error('Error fetching wallet balance:', error);
                updateWalletDisplay(0); // Default to 0 if there's an error
            });
    }

    // Function to update wallet display based on balance
    function updateWalletDisplay(balance) {

        // Update the wallet balance text
        walletBalanceDisplay.textContent = `Wallet Balance ₹ ${balance.toFixed(2)}${balance < total ? ' Unavailable' : ''}`;

        // Check if wallet balance is sufficient
        if (balance < total) {
            // Disable wallet payment option - add 'disabled' class to match COD style
            walletPaymentOption.classList.add('disabled');
            walletPaymentOption.style.cursor = 'not-allowed';

            // Update message to show insufficient balance
            walletMessageDisplay.innerHTML = `
        <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
            <path d="M10 18a8 8 0 100-16 8 8 0 000 16zM9 9a1 1 0 112 0v4a1 1 0 11-2 0V9z"></path>
        </svg>
        Insufficient balance.
    `;

            // Show gift card input when balance is less than total
            if (walletInput) {
                walletInput.classList.add('active');

                // Force show the input by setting inline styles
                walletInput.style.maxHeight = '200px';
                walletInput.style.opacity = '1';
                walletInput.style.transform = 'translateY(0)';

            }

            // If this was selected, choose another payment method
            if (selectedPaymentMethod === 'Wallet') {
                // Find the first available payment method
                const availableMethod = document.querySelector('.payment-option:not(.disabled)');
                if (availableMethod) {
                    availableMethod.click();
                }
            }
        } else {
            // Enable wallet payment option
            walletPaymentOption.classList.remove('disabled');
            walletPaymentOption.style.cursor = 'pointer';

            // Update message to show wallet is available
            walletMessageDisplay.innerHTML = `
        Available 
    `;

            // Hide gift card input if balance is sufficient
            if (walletInput) {
                walletInput.classList.remove('active');

                // Reset inline styles
                walletInput.style.maxHeight = '';
                walletInput.style.opacity = '';
                walletInput.style.transform = '';
            }
        }
    }

    // Check COD availability
    if (!isCodAvailable) {
        // Disable COD payment option
        codPaymentOption.classList.add('disabled');
        codPaymentOption.style.cursor = 'not-allowed';
    }

    // Apply gift card functionality
    // Apply gift card functionality
    applyButton.addEventListener('click', function () {
        const giftCardCode = giftCardInput.value.trim();
        const responseDiv = document.getElementById('gift-card-response');

        // Clear previous message
        responseDiv.classList.remove('text-green-600', 'text-red-600');
        responseDiv.classList.add('hidden');
        responseDiv.textContent = '';

        if (!giftCardCode) {
            responseDiv.textContent = 'Please enter a gift card code';
            responseDiv.classList.add('text-red-600', 'block');
            responseDiv.classList.remove('hidden');
            return;
        }

        // Show loading state
        this.innerHTML = '<i class="fas fa-spinner fa-spin"></i>';
        this.disabled = true;

        // Send AJAX request to apply gift card
        fetch('/checkout/redeem/gift/code', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                code: giftCardCode
            })
        })
            .then(response => {
                if (!response.ok) {
                    return response.json().then(err => { throw new Error(err.message || 'Invalid gift card code'); });
                }
                return response.json();
            })
            .then(data => {
                // Show success message
                responseDiv.textContent = data.message || 'Gift card applied successfully';
                responseDiv.classList.add('text-green-600', 'block');
                responseDiv.classList.remove('hidden');

                // Clear input field
                giftCardInput.value = '';

                // Fetch updated wallet balance
                fetchWalletBalance();
            })
            .catch(error => {
                responseDiv.textContent = error.message || 'Failed to apply gift card';
                responseDiv.classList.add('text-red-600', 'block');
                responseDiv.classList.remove('hidden');
            })
            .finally(() => {
                // Reset button state
                this.innerHTML = 'Apply';
                this.disabled = false;
            });
    });

    // Set up event listeners for each payment option
    paymentOptions.forEach(option => {
        option.addEventListener('click', function () {
            // Check if this option is disabled
            if (this.classList.contains('disabled')) {
                return;
            }

            // Remove selected class from all options
            paymentOptions.forEach(opt => {
                opt.classList.remove('selected');
                opt.classList.remove('just-selected');
                opt.querySelector('.custom-radio').classList.remove('selected');
            });

            // Add selected class to clicked option
            this.classList.add('selected');
            this.classList.add('just-selected');
            this.querySelector('.custom-radio').classList.add('selected');

            // Store the selected payment method
            selectedPaymentMethod = this.getAttribute('data-value');

            // Toggle wallet input visibility
            if (this.id === 'walletPayment') {
                const walletBalance = parseFloat(walletBalanceDisplay.textContent.match(/₹\s+([\d.]+)/)[1]);

                if (walletBalance < total) {
                    walletInput.classList.add('active');

                    // Force show the input
                    walletInput.style.maxHeight = '200px';
                    walletInput.style.opacity = '1';
                    walletInput.style.transform = 'translateY(0)';

                } else {
                    walletInput.classList.remove('active');

                    // Reset inline styles
                    walletInput.style.maxHeight = '';
                    walletInput.style.opacity = '';
                    walletInput.style.transform = '';
                }
            } else {
                walletInput.classList.remove('active');

                // Reset inline styles
                walletInput.style.maxHeight = '';
                walletInput.style.opacity = '';
                walletInput.style.transform = '';
            }

            // Remove the pulse animation class after animation completes
            setTimeout(() => {
                this.classList.remove('just-selected');
            }, 800);
        });
    });

    // Add hover effects
    paymentOptions.forEach(option => {
        option.addEventListener('mouseenter', function () {
            if (!this.classList.contains('selected') && !this.classList.contains('disabled')) {
                this.style.backgroundColor = '#f9fafb';
            }
        });

        option.addEventListener('mouseleave', function () {
            if (!this.classList.contains('selected')) {
                this.style.backgroundColor = '';
            }
        });
    });

    // Helper functions for showing toasts
    function showErrorToast(message) {
        console.error('Error:', message);
        // Add your toast implementation here
    }

    function showSuccessToast(message) {
        console.log('Success:', message);
        // Add your toast implementation here
    }

    // Set default selection (first available payment method)
    const availableMethod = document.querySelector('.payment-option:not(.disabled)');
    if (availableMethod) {
        availableMethod.click();
    }

    // Add tooltips to disabled payment options
    const disabledOptions = document.querySelectorAll('.payment-option.disabled');

    // Add tooltip elements to each disabled option
    disabledOptions.forEach(option => {
        // Extract the reason for disability
        let reasonText = "This payment option is unavailable";

        // Get specific messages
        if (option.id === 'walletPayment') {
            reasonText = "Insufficient wallet balance";
        } else if (option.id === 'codPayment') {
            reasonText = "Cash on Delivery is not available for this order";
        }

        // Create tooltip
        const tooltip = document.createElement('div');
        tooltip.className = 'disabled-tooltip';
        tooltip.textContent = reasonText;
        option.appendChild(tooltip);

        // Add shake effect when clicked
        option.addEventListener('click', function () {
            if (this.classList.contains('disabled')) {
                this.classList.add('shake');
                setTimeout(() => {
                    this.classList.remove('shake');
                }, 500);
            }
        });
    });

    // Force check the wallet balance and update UI after a small delay
    setTimeout(() => {
        const walletBalance = parseFloat(walletBalanceDisplay.textContent.match(/₹\s+([\d.]+)/)[1]) || 0;

        if (walletBalance < total) {
            walletInput.classList.add('active');

            // Force show the input
            walletInput.style.maxHeight = '200px';
            walletInput.style.opacity = '1';
            walletInput.style.transform = 'translateY(0)';
        }
    }, 100);
});

// Process payment function
document.getElementById('proceedToPay').addEventListener('click', function () {
    console.log('Selected Payment Method:', selectedPaymentMethod);
    console.log('Address ID:', document.getElementById('addressId').value);
    let addressIdElement = document.getElementById('addressId');
    let couponIDElement = document.getElementById('couponID');
    let couponCodeElement = document.getElementById('couponCode');
    let couponDiscountElement = document.getElementById('couponDiscount');

    if (!selectedPaymentMethod || !addressIdElement || !addressIdElement.value) {
        showErrorToast("Missing payment method or address.");
        return;
    }

    let addressId = addressIdElement.value;
    let couponID = couponIDElement.value;
    let couponCode = couponCodeElement.value;
    let couponDiscount = couponDiscountElement.value;

    // Show loading state
    this.innerHTML = '<i class="fas fa-spinner fa-spin mr-2"></i> Processing...';
    this.disabled = true;

    fetch('/checkout/payment/proceed', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            paymentMethod: selectedPaymentMethod,
            addressId: addressId,
            couponCode: couponCode,
            couponId: couponID,
            couponDiscountAmount: couponDiscount
        }),
    })
        .then(response => {
            if (!response.ok) {
                return response.json().then(err => { throw new Error(err.message || "Payment failed."); });
            }
            return response.json(); // Expecting JSON response
        })
        .then(data => {
            if (selectedPaymentMethod === 'Razorpay') {
                initializeRazorpay(data);
            }
            else if (selectedPaymentMethod === 'COD') {
                window.location.href = '/order/success';
            }
            else if (selectedPaymentMethod === 'Wallet') {
                window.location.href = '/order/success';
            }
        })
        .catch(error => {
            console.error('Error:', error);
            showErrorToast(error.message || "An error occurred while processing the payment.");

            // Reset button state
            this.innerHTML = '<span>Proceed to Pay</span><i class="fas fa-arrow-right ml-2"></i>';
            this.disabled = false;
        });
});

// Initialize Razorpay handler
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
            // Handle successful payment
            verifyPayment({
                razorpay_payment_id: response.razorpay_payment_id,
                razorpay_order_id: response.razorpay_order_id,
                razorpay_signature: response.razorpay_signature,
            });
        },
        prefill: {
            name: data.prefill.name,
            email: data.prefill.email,
            contact: data.prefill.contact
        },
        notes: data.notes,
        theme: {
            color: "#0000"
        },
        modal: {
            ondismiss: function () {
                const payButton = document.getElementById('proceedToPay');
                payButton.innerHTML = '<span>Proceed to Pay</span><i class="fas fa-arrow-right ml-2"></i>';
                payButton.disabled = false;
                showErrorToast("Payment cancelled");
            }
        },
        // Add retry configuration to prevent retry option
        retry: {
            enabled: false  // Disable retry functionality
        }
    };

    const rzp = new Razorpay(options);

    // Add payment failed handler
    rzp.on('payment.failed', function (response) {
        fetch('/order/failed', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                razorpay_payment_id: response.error.metadata.payment_id,
                razorpay_order_id: response.error.metadata.order_id
            })
        })
            .then(response => {
                if (response.redirected) {
                    window.location.href = response.url; // Manually redirect browser
                } else {
                    return response.json();
                }
            })
            .then(data => {
            })
            .catch(error => console.error('Error:', error));

        // Optional: Delay before redirection to show failure message
        setTimeout(function () {
            window.location.href = '/payment-failed';
        }, 3000);
    });

    rzp.open();
}

// Function to verify payment with backend
function verifyPayment(paymentData) {
    fetch('/checkout/payment/verify', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(paymentData),
    })
        .then(response => {
            if (!response.ok) {
                return response.json().then(err => { throw new Error(err.message || "Payment verification failed."); });
            }
            return response.json();
        })
        .then(data => {
            // Redirect to success page
            window.location.href = '/order/success';
        })
        .catch(error => {
            console.error('Error:', error);
            showErrorToast(error.message || "An error occurred while verifying the payment.");

            // Reset button state
            const payButton = document.getElementById('proceedToPay');
            payButton.innerHTML = '<span>Proceed to Pay</span><i class="fas fa-arrow-right ml-2"></i>';
            payButton.disabled = false;
        });
}