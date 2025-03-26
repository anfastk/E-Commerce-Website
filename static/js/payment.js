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

    // Initial wallet input state
    walletInput.style.display = 'block';
    walletInput.style.transition = 'all 0.3s ease';

    // Fetch wallet balance initially
    fetchWalletBalance();

    function fetchWalletBalance() {
        fetch('/checkout/check/wallet/balance', {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json'
            }
        })
            .then(response => response.json())
            .then(data => {
                updateWalletDisplay(data.balance);
            })
            .catch(error => {
                console.error('Error fetching wallet balance:', error);
                updateWalletDisplay(0);
            });
    }

    function updateWalletDisplay(balance) {
        walletBalanceDisplay.textContent = `Wallet Balance ₹ ${balance.toFixed(2)}${balance < total ? ' Unavailable' : ''}`;

        const isInsufficient = balance < total;

        if (isInsufficient) {
            walletPaymentOption.classList.add('disabled');
            walletPaymentOption.style.cursor = 'not-allowed';
            walletMessageDisplay.innerHTML = `
                <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                    <path d="M10 18a8 8 0 100-16 8 8 0 000 16zM9 9a1 1 0 112 0v4a1 1 0 11-2 0V9z"></path>
                </svg>
                Insufficient balance.
            `;

            // Show wallet input when balance is insufficient
            walletInput.style.maxHeight = '200px';
            walletInput.style.opacity = '1';
            walletInput.style.transform = 'translateY(0)';
        } else {
            walletPaymentOption.classList.remove('disabled');
            walletPaymentOption.style.cursor = 'pointer';
            walletMessageDisplay.innerHTML = `Available`;

            // Hide wallet input when balance is sufficient
            walletInput.style.maxHeight = '0';
            walletInput.style.opacity = '0';
            walletInput.style.transform = 'translateY(-10px)';
        }

        // Update selected payment method if necessary
        if (selectedPaymentMethod === 'Wallet' && isInsufficient) {
            const availableMethod = document.querySelector('.payment-option:not(.disabled)');
            if (availableMethod) availableMethod.click();
        }
    }

    // Check COD availability
    if (!isCodAvailable) {
        codPaymentOption.classList.add('disabled');
        codPaymentOption.style.cursor = 'not-allowed';
    }

    // Apply gift card functionality (unchanged)
    applyButton.addEventListener('click', function () {
        const giftCardCode = giftCardInput.value.trim();
        const responseDiv = document.getElementById('gift-card-response');

        responseDiv.classList.remove('text-green-600', 'text-red-600');
        responseDiv.classList.add('hidden');
        responseDiv.textContent = '';

        if (!giftCardCode) {
            responseDiv.textContent = 'Please enter a gift card code';
            responseDiv.classList.add('text-red-600', 'block');
            responseDiv.classList.remove('hidden');
            return;
        }

        this.innerHTML = '<i class="fas fa-spinner fa-spin"></i>';
        this.disabled = true;

        fetch('/checkout/redeem/gift/code', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ code: giftCardCode })
        })
            .then(response => {
                if (!response.ok) {
                    return response.json().then(err => { throw new Error(err.message || 'Invalid gift card code'); });
                }
                return response.json();
            })
            .then(data => {
                responseDiv.textContent = data.message || 'Gift card applied successfully';
                responseDiv.classList.add('text-green-600', 'block');
                responseDiv.classList.remove('hidden');
                giftCardInput.value = '';
                fetchWalletBalance();
            })
            .catch(error => {
                responseDiv.textContent = error.message || 'Failed to apply gift card';
                responseDiv.classList.add('text-red-600', 'block');
                responseDiv.classList.remove('hidden');
            })
            .finally(() => {
                this.innerHTML = 'Apply';
                this.disabled = false;
            });
    });

    // Payment option selection
    paymentOptions.forEach(option => {
        option.addEventListener('click', function () {
            if (this.classList.contains('disabled')) return;

            paymentOptions.forEach(opt => {
                opt.classList.remove('selected', 'just-selected');
                opt.querySelector('.custom-radio').classList.remove('selected');
            });

            this.classList.add('selected', 'just-selected');
            this.querySelector('.custom-radio').classList.add('selected');
            selectedPaymentMethod = this.getAttribute('data-value');

            const walletBalance = parseFloat(walletBalanceDisplay.textContent.match(/₹\s+([\d.]+)/)[1]) || 0;

            // Handle wallet input visibility
            if (this.id === 'walletPayment') {
                if (walletBalance < total) {
                    walletInput.style.maxHeight = '200px';
                    walletInput.style.opacity = '1';
                    walletInput.style.transform = 'translateY(0)';
                } else {
                    walletInput.style.maxHeight = '0';
                    walletInput.style.opacity = '0';
                    walletInput.style.transform = 'translateY(-10px)';
                }
            } else {
                // Show wallet input for other payment methods only if balance < total
                if (walletBalance < total) {
                    walletInput.style.maxHeight = '200px';
                    walletInput.style.opacity = '1';
                    walletInput.style.transform = 'translateY(0)';
                } else {
                    walletInput.style.maxHeight = '0';
                    walletInput.style.opacity = '0';
                    walletInput.style.transform = 'translateY(-10px)';
                }
            }

            setTimeout(() => this.classList.remove('just-selected'), 800);
        });
    });

    // Add hover effects (unchanged)
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

    // Set default selection
    const availableMethod = document.querySelector('.payment-option:not(.disabled)');
    if (availableMethod) availableMethod.click();

    // Add tooltips to disabled options (unchanged)
    const disabledOptions = document.querySelectorAll('.payment-option.disabled');
    disabledOptions.forEach(option => {
        let reasonText = option.id === 'walletPayment' ?
            "Insufficient wallet balance" :
            "Cash on Delivery is not available for this order";

        const tooltip = document.createElement('div');
        tooltip.className = 'disabled-tooltip';
        tooltip.textContent = reasonText;
        option.appendChild(tooltip);

        option.addEventListener('click', function () {
            if (this.classList.contains('disabled')) {
                this.classList.add('shake');
                setTimeout(() => this.classList.remove('shake'), 500);
            }
        });
    });
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
            // Check the payment method to handle response appropriately
            if (selectedPaymentMethod === 'Razorpay') {
                return response.json(); // Razorpay still expects JSON
            } else {
                return response.text(); // COD and Wallet expect HTML
            }
        })
        .then(data => {
            if (selectedPaymentMethod === 'Razorpay') {
                initializeRazorpay(data); // Handle Razorpay as before
            } else if (selectedPaymentMethod === 'COD' || selectedPaymentMethod === 'Wallet') {
                // Instead of redirecting, render the HTML response
                document.open();
                document.write(data); // Write the HTML response to the current window
                document.close();
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
            return response.text(); // Expecting HTML response instead of JSON
        })
        .then(data => {
            // Render the HTML response instead of redirecting
            document.open();
            document.write(data); // Write the HTML response to the current window
            document.close();
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