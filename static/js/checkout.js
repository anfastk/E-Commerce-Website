// Global variable to track selected address
let selectedAddressId = null;
 
// Address Management Functions
async function loadAddresses() {
    try {
        const response = await fetch('/checkout/addresses', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' }
        });
        const data = await response.json();

        if (data.status === 'Success') {
            const container = document.getElementById('address-container');
            container.innerHTML = '';

            document.getElementById('address-subtitle').style.display =
                data.Addresses.length >= 2 ? 'block' : 'none';

            data.Addresses.forEach(address => {
                const isDefault = address.is_default;
                container.innerHTML += `
            <div class="address-card border rounded-lg p-4 ${isDefault ? 'selected' : ''}" 
                 data-address-id="${address.ID}"
                 onclick="handleCardClick(this, event)">
                <div class="flex items-start gap-4">
                    <input type="radio" name="address_card" class="mt-1 cursor-pointer" 
                        value="${address.ID}" ${isDefault ? 'checked' : ''} 
                        onclick="event.stopPropagation(); selectAddress('${address.ID}', this.parentElement.parentElement)">
                    <div class="flex-grow">
                        <h3 class="font-medium">${address.user_firstname} ${address.user_lastname}</h3>
                        <p class="text-sm text-gray-600">
                            ${address.user_address} <br> 
                            ${address.user_city}, ${address.user_state} ${address.user_pincode} <br> 
                            ${address.user_country}
                        </p>
                        <p class="text-sm text-gray-600">LandMark: ${address.user_landmark}</p>
                    </div>
                </div>
                <div class="action-buttons flex gap-4 mt-4 ml-8" onclick="event.stopPropagation()">
                    <button onclick="openEditModal('${address.ID}', '${address.user_firstname}', 
                        '${address.user_lastname}', '${address.user_landmark}', '${address.user_address}', 
                        '${address.user_country}', '${address.user_state}', '${address.user_city}', 
                        '${address.user_pincode}', '${address.user_number}')"
                        class="text-gray-600 flex items-center gap-1 hover:text-gray-800">
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                                d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                        </svg>
                        Edit
                    </button>
                    <button onclick="deleteAddress('${address.ID}')"
                        class="text-red-400 flex items-center gap-1 hover:text-red-600">
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                                d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                        </svg>
                        Delete
                    </button>
                </div>
            </div>
        `;

                if (isDefault) {
                    selectedAddressId = address.ID;
                }
            });

            if (!selectedAddressId && data.Addresses.length > 0) {
                const firstAddress = data.Addresses[0];
                selectedAddressId = firstAddress.ID;
                const firstCard = container.querySelector('.address-card');
                firstCard.classList.add('selected');
                firstCard.querySelector('input[type="radio"]').checked = true;
            }
        }
    } catch (error) {
        console.error('Error loading addresses:', error);
        showErrorToast('Failed to load addresses');
    }
}

async function handleAddAddress(e) {
    e.preventDefault();
    const formData = new FormData(e.target);
    const data = Object.fromEntries(formData.entries());

    try {
        const response = await fetch('/profile/add/address', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data)
        });

        if (response.ok) {
            closeModal('addAddressModal');
            document.getElementById('addAddressForm').reset();
            showSuccessToast('Address added successfully');
            await loadAddresses();
        } else {
            throw new Error('Failed to add address');
        }
    } catch (error) {
        console.error('Error:', error);
        showErrorToast('Failed to add address');
    }
}

function openEditModal(id, firstName, lastName, landmark, address, country, state, city, pinCode, mobile) {
    const form = document.getElementById('editAddressForm');
    form.id.value = id;
    form.firstName.value = firstName;
    form.lastName.value = lastName;
    form.landmark.value = landmark;
    form.address.value = address;
    form.country.value = country;
    form.state.value = state;
    form.city.value = city;
    form.zipCode.value = pinCode;
    form.phoneNumber.value = mobile;
    openModal('editAddressModal');
}

async function handleEditAddress(e) {
    e.preventDefault();
    const formData = new FormData(e.target);
    const data = Object.fromEntries(formData.entries());

    try {
        const response = await fetch('/profile/edit/address', {
            method: 'PATCH',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data)
        });

        if (response.ok) {
            closeModal('editAddressModal');
            showSuccessToast('Address updated successfully');
            await loadAddresses();
        } else {
            throw new Error('Failed to update address');
        }
    } catch (error) {
        console.error('Error:', error);
        showErrorToast('Failed to update address');
    }
}

async function deleteAddress(id) {
    if (!confirm('Are you sure you want to delete this address?')) return;

    try {
        const response = await fetch(`/profile/delete/address/${id}`, {
            method: 'DELETE'
        });

        if (response.ok) {
            showSuccessToast('Address deleted successfully');
            await loadAddresses();
        } else {
            throw new Error('Failed to delete address');
        }
    } catch (error) {
        console.error('Error:', error);
        showErrorToast('Failed to delete address');
    }
}

function openModal(modalId) {
    document.getElementById(modalId).classList.add('active');
    document.body.style.overflow = 'hidden';
}

function closeModal(modalId) {
    document.getElementById(modalId).classList.remove('active');
    document.body.style.overflow = 'auto';
}

function selectAddress(addressId, card) {
    selectedAddressId = addressId;
    document.querySelectorAll('.address-card').forEach(c => {
        c.classList.remove('selected');
        const radio = c.querySelector('input[type="radio"]');
        if (radio) radio.checked = false;
    });
    card.classList.add('selected');
    const radioInput = card.querySelector('input[type="radio"]');
    if (radioInput) radioInput.checked = true;
}

function handleCardClick(card, event) {
    if (event.target.closest('.action-buttons')) return;
    const addressId = card.getAttribute('data-address-id');
    selectAddress(addressId, card);
}

document.getElementById('addAddressForm').addEventListener('submit', handleAddAddress);
document.getElementById('editAddressForm').addEventListener('submit', handleEditAddress);

document.querySelectorAll('.modal').forEach(modal => {
    modal.addEventListener('click', function (e) {
        if (e.target === this) closeModal(this.id);
    });
});

document.addEventListener('DOMContentLoaded', loadAddresses);

// Order Summary and Coupon Logic
document.getElementById('proceedToPaymentBtn').addEventListener('click', async function (e) {
    e.preventDefault();

    // Re-check the selected radio button to ensure selectedAddressId is current
    const selectedRadio = document.querySelector('input[name="address_card"]:checked');
    if (selectedRadio) {
        selectedAddressId = selectedRadio.value;
    }

    if (!selectedAddressId) {
        showErrorToast('Please select a shipping address');
        return;
    }


    const paymentData = {
        addressId: selectedAddressId
    };

    if (appliedCoupon) {
        paymentData.couponCode = appliedCoupon.code;
        paymentData.couponId = appliedCoupon.couponId;
        paymentData.couponDiscountAmount = couponDiscountAmount;
    }


    try {
        const response = await fetch('/checkout/payment', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Accept': 'application/json'
            },
            body: JSON.stringify(paymentData),
        });

        if (response.ok) {
            const contentType = response.headers.get('content-type');
            if (contentType && contentType.includes('application/json')) {
                const result = await response.json();
                window.location.href = result.redirectUrl;
            } else {
                const result = await response.text();
                document.open();
                document.write(result);
                document.close();
            }
        } else {
            const result = await response.json();
            throw new Error(result.message || 'Payment proceeding failed');
        }
    } catch (error) {
        console.error('Payment Error:', error);
        showErrorToast(error.message || 'Error proceeding to payment');
    }
});