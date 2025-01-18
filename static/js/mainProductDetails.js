const addOfferBtn = document.getElementById("addOfferBtn");
const offerModal = document.getElementById("offerModal");
const closeAddOfferModal = document.getElementById("closeAddOfferModal");

if (addOfferBtn) {
    addOfferBtn.addEventListener("click", () => {
        offerModal.classList.remove("hidden"); // Modal visible aakkan hidden remove cheyyunnu
    });
}

if (closeAddOfferModal) {
    closeAddOfferModal.addEventListener("click", () => {
        offerModal.classList.add("hidden"); // Modal hide cheyyunnu
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

optionsBtn.addEventListener("click", function (event) {
    event.stopPropagation(); // Prevent the click from bubbling up to the window
    optionsMenu.classList.toggle("hidden");
});

window.addEventListener("click", function (event) {
    if (!event.target.closest(".group")) {
        optionsMenu.classList.add("hidden");
    }
});