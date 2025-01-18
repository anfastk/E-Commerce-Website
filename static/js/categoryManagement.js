document.addEventListener('DOMContentLoaded', () => {
  // Add event listeners to all buttons with the "edit-btn" class
  document.querySelectorAll('.edit-btn').forEach(button => {
    button.addEventListener('click', () => {
      const editPopup = document.getElementById('edit-popup');
      if (editPopup) {
        editPopup.classList.remove('hidden'); // Show the popup
      } else {
        console.error("Element with ID 'edit-popup' not found.");
      }
    });
  });

  // Close button logic for the edit popup
  const closeEditButton = document.getElementById('close-edit-popup');
  if (closeEditButton) {
    closeEditButton.addEventListener('click', () => {
      const editPopup = document.getElementById('edit-popup');
      if (editPopup) {
        editPopup.classList.add('hidden'); // Hide the popup
      } else {
        console.error("Element with ID 'edit-popup' not found.");
      }
    });
  } else {
    console.error("Close button with ID 'close-edit-popup' not found.");
  }

  // Add Category Popup Logic
  const toggleModal = () => {
    const modal = document.getElementById('categoryModal');
    if (modal) {
      modal.classList.toggle('hidden'); // Toggle the popup visibility
    } else {
      console.error("Element with ID 'categoryModal' not found.");
    }
  };

  const addCategoryButton = document.querySelector('[onclick="toggleModal()"]');
  if (addCategoryButton) {
    addCategoryButton.addEventListener('click', () => {
      toggleModal();
    });
  } else {
    console.error("Add Category button with onclick='toggleModal()' not found.");
  }

  const closeCategoryButton = document.getElementById('close-category-popup');
  if (closeCategoryButton) {
    closeCategoryButton.addEventListener('click', () => {
      toggleModal(); // Close the Add Category popup
    });
  } else {
    console.error("Close button with ID 'close-category-popup' not found.");
  }
});
document.addEventListener('DOMContentLoaded', () => {
  // Add event listener for the "Manage Offer" button
  const manageOfferButton = document.getElementById('manage-offer-btn');
  if (manageOfferButton) {
    manageOfferButton.addEventListener('click', () => {
      const manageOfferPopup = document.getElementById('manage-offer-popup');
      if (manageOfferPopup) {
        manageOfferPopup.classList.toggle('hidden'); // Toggle visibility of the popup
      } else {
        console.error("Element with ID 'manage-offer-popup' not found.");
      }
    });
  } else {
    console.error("Button with ID 'manage-offer-btn' not found.");
  }

  // Close button logic for the "Manage Offer" popup
  const closeManageOfferButton = document.getElementById('close-manage-offer-popup');
  if (closeManageOfferButton) {
    closeManageOfferButton.addEventListener('click', () => {
      const manageOfferPopup = document.getElementById('manage-offer-popup');
      if (manageOfferPopup) {
        manageOfferPopup.classList.add('hidden'); // Hide the popup
      } else {
        console.error("Element with ID 'manage-offer-popup' not found.");
      }
    });
  } else {
    console.error("Close button with ID 'close-manage-offer-popup' not found.");
  }
});
