// Modal Functionality
const openPopup = document.getElementById("openPopup");
const popupModal = document.getElementById("popupModal");
const closePopup = document.getElementById("closePopup");

openPopup.addEventListener("click", () => {
    popupModal.classList.remove("hidden");
});

closePopup.addEventListener("click", () => {
    popupModal.classList.add("hidden");
});

// Add More Specification Functionality
const addSpecBtn = document.getElementById("addSpecBtn");
const specificationsList = document.getElementById("specificationsList");

addSpecBtn.addEventListener("click", () => {
    const newSpec = document.createElement("div");
    newSpec.classList.add("specification-item", "flex", "flex-row", "gap-4", "mt-4");
    newSpec.innerHTML = `
           <div class="flex-1">
               <label for="key" class="block text-gray-700 font-medium">Specification Key</label>
               <input type="text" name="key[]" class="w-full p-2 border border-gray-300 rounded-lg" placeholder="e.g., Material" required>
           </div>
           <div class="flex-1">
               <label for="value" class="block text-gray-700 font-medium">Specification Value</label>
               <input type="text" name="value[]" class="w-full p-2 border border-gray-300 rounded-lg" placeholder="e.g., Cotton" required>
           </div>
       `;
    specificationsList.appendChild(newSpec);
});

 // Add click event listener for all option buttons
 document.querySelectorAll('.options-btn').forEach((btn) => {
    btn.addEventListener('click', function () {
        // Find the sibling dropdown menu
        const dropdown = this.nextElementSibling;

        // Toggle the hidden class
        dropdown.classList.toggle('hidden');
    });
});

// Optional: Close the dropdown when clicking outside
document.addEventListener('click', function (event) {
    const isClickInside = event.target.closest('.group');
    if (!isClickInside) {
        document.querySelectorAll('.options').forEach((dropdown) => {
            dropdown.classList.add('hidden');
        });
    }
});