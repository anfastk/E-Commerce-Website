const addOfferBtn = document.getElementById("addOfferBtn");
const offerModal = document.getElementById("offerModal");
const closeAddOfferModal = document.getElementById("closeAddOfferModal");

if (addOfferBtn) {
    addOfferBtn.addEventListener("click", () => {
        offerModal.classList.remove("hidden");
    });
}

if (closeAddOfferModal) {
    closeAddOfferModal.addEventListener("click", () => {
        offerModal.classList.add("hidden");
    });
}

// Update/Delete Offer Modal
const updateModal = document.getElementById("updateModal");
const closeUpdateModal = document.getElementById("closeUpdateModal");
const deleteOfferBtn = document.getElementById("deleteOfferBtn");
const updateOfferBtn = document.getElementById("updateOfferBtn");

updateOfferBtn.addEventListener("click", () => {
    updateModal.classList.remove("hidden"); // Show the modal by removing the 'hidden' class
});

closeUpdateModal.addEventListener("click", () => {
    updateModal.classList.add("hidden");
});

deleteOfferBtn.addEventListener("click", () => {
    const confirmation = confirm("Are you sure you want to delete this offer?");

});

// Open/Close Description Modal
const openModal = document.getElementById("openModal");
const closeDescriptionModal = document.getElementById("closeModal"); // Close modal element
const modal = document.getElementById("descriptionModal");

openModal.onclick = () => {
    modal.classList.remove("hidden");
};

closeDescriptionModal.onclick = () => {
    modal.classList.add("hidden");
};

// Add a new pair (heading and description)
const addPairButton = document.getElementById("addPair");
const keyValuePairs = document.getElementById("keyValuePairs");

addPairButton.onclick = () => {
    const newPair = document.createElement('div');
    newPair.classList.add('flex', 'mb-4', 'space-x-4');
    newPair.innerHTML = `
        <input type="text" name="heading[]" placeholder="Enter heading" class="p-2 border border-gray-300 rounded w-1/3" required>
        <textarea name="description[]" placeholder="Enter description" class="p-2 border border-gray-300 rounded w-2/3" required></textarea>
    `;
    keyValuePairs.appendChild(newPair);
};


const optionsBtn = document.getElementById("options-btn");
const optionsMenu = document.getElementById("options");

if (optionsBtn && optionsMenu) {
    optionsBtn.addEventListener("click", function (event) {
        event.stopPropagation(); 
        optionsMenu.classList.toggle("hidden");
    });

    window.addEventListener("click", function (event) {
        if (!event.target.closest(".group")) {
            optionsMenu.classList.add("hidden");
        }
    });
}

window.addEventListener("click", function (event) {
    if (!event.target.closest(".group")) {
        optionsMenu.classList.add("hidden");
    }
});

document.addEventListener('DOMContentLoaded', () => {
    const openPopupBtn = document.getElementById('openUpdatePopup');
    const closePopupBtn = document.getElementById('closeUpdatePopup');
    const popupModal = document.getElementById('updatePopupModal');
    const updateDescriptionsForm = document.getElementById('updateDescriptionsForm');

    // Open popup
    openPopupBtn.addEventListener('click', () => {
        popupModal.classList.remove('hidden');
    });

    // Close popup
    closePopupBtn.addEventListener('click', () => {
        popupModal.classList.add('hidden');
    });

    // Close popup if clicking outside the modal
    popupModal.addEventListener('click', (event) => {
        if (event.target === popupModal) {
            popupModal.classList.add('hidden');
        }
    });

    // Form submission handling
    updateDescriptionsForm.addEventListener('submit', async (event) => {
        event.preventDefault();

        // Create FormData object
        const formData = new FormData(updateDescriptionsForm);

        // Convert FormData to an object
        const data = {
            description_id: formData.getAll('description_id[]'),
            heading: formData.getAll('heading[]'),
            description: formData.getAll('description[]')
        };

        try {
            const response = await fetch(updateDescriptionsForm.action, {
                method: 'PATCH', // Change back to POST
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(data)
            });

            const responseData = await response.json();

            if (response.ok) {
                popupModal.classList.add('hidden');
                // Optionally, you can add a success message or reload the page
                window.location.reload();
            } else {
                // Handle error
                console.error('Submission failed', responseData);
                alert(responseData.error || 'Failed to update descriptions');
            }
        } catch (error) {
            console.error('Error:', error);
            alert('An error occurred while updating descriptions');
        }
    });
});

function deleteDescription(descriptionId, productId) {
    // Split the IDs if they're passed as a comma-separated string
    const ids = descriptionId.split(',');
    const descId = ids[0];
    const prodId = ids[1] || productId;

    fetch(`/admin/products/variant/description/delete/${descId}`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        }
    }).then(response => {
        if (response.ok) {
            // Remove the description item from the DOM
            const descriptionItem = document.querySelector(`.Descriptions-item[data-desc-id="${descId}"]`);
            if (descriptionItem) {
                descriptionItem.remove();
            }
            // Redirect to product details page
            window.location.href = `/admin/products/main/details?product_id=${prodId}`;
        } else {
            // Handle error cases
            console.error('Failed to delete description');
        }
    }).catch(error => {
        console.error('Error:', error);
    });
}