let cropper = null;
let currentImageElement = null;
let currentFile = null;
const CROP_WIDTH = 400;
const CROP_HEIGHT = 400;

// Open Upload Popup
document.getElementById('openUploadPopup').addEventListener('click', function () {
    document.getElementById('imageUploadPopup').classList.remove('hidden');
});

// Close Upload Popup
function closeUploadPopup() {
    document.getElementById('imageUploadPopup').classList.add('hidden');
    document.getElementById('banner-preview').innerHTML = '';
}
 
function confirmUpload() {
    const previewImg = document.getElementById('banner-preview').querySelector('img');
    const productId = document.getElementById('product-id').value;

    if (!previewImg) {
        showErrorToast('Please upload an image first');
        return;
    }

    // Show loading toast
    showSuccessToast('Uploading image...');

    // Convert image to file
    fetch(previewImg.src)
        .then(res => res.blob())
        .then(blob => {
            const formData = new FormData();
            formData.append('product_image', blob, 'uploaded-image.png');
            formData.append('product_id', productId);

            fetch('/admin/products/main/image/change', {
                method: 'POST',
                body: formData
            })
                .then(response => response.json())
                .then(data => {
                    if (data.filename) {
                        showSuccessToast('Image uploaded successfully');
                        closeUploadPopup();
                        setTimeout(() => {
                            window.location.reload();
                        }, 1000);
                    } else {
                        showErrorToast('Upload failed');
                    }
                })
                .catch(error => {
                    console.error('Upload error:', error);
                    showErrorToast('Failed to upload image');
                });
        });
}

// Drag and Drop Functionality
function enableDragAndDrop() {
    const dropArea = document.getElementById('banner-drop-area');
    const input = document.getElementById('banner-input');
    const previewContainer = document.getElementById('banner-preview');

    ['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => {
        dropArea.addEventListener(eventName, preventDefaults, false);
    });

    function preventDefaults(e) {
        e.preventDefault();
        e.stopPropagation();
    }

    dropArea.addEventListener('drop', (e) => {
        const dt = e.dataTransfer;
        const files = dt.files;
        handleFileUpload(files);
    });

    input.addEventListener('change', (e) => {
        handleFileUpload(e.target.files);
    });
}

function handleFileUpload(files) {
    const previewContainer = document.getElementById('banner-preview');
    previewContainer.innerHTML = ''; // Clear previous previews

    if (files.length > 1) {
        showErrorToast('Please upload only one image');
        return;
    }

    const file = files[0];
    if (!file.type.startsWith('image/')) {
        showErrorToast('Please upload only image files');
        return;
    }

    const reader = new FileReader();
    reader.onload = function (e) {
        const preview = document.createElement('div');
        preview.className = 'relative border rounded p-2';
        preview.innerHTML = `
                <img src="${e.target.result}" alt="" class="w-full h-40 object-contain">
                <div class="absolute top-2 right-2 flex gap-2">
                    <button type="button" class="bg-blue-500 text-white p-1 rounded" 
                        onclick="startCrop(this.parentElement.parentElement.querySelector('img'), '${file.name}')">
                        Crop
                    </button>
                </div>
            `;
        previewContainer.appendChild(preview);
        showSuccessToast('Image loaded successfully');
    };

    reader.onerror = function () {
        showErrorToast('Failed to load image');
    };

    reader.readAsDataURL(file);
}

function startCrop(imgElement, fileName) {
    const modal = document.getElementById('cropModal');
    const cropImage = document.getElementById('cropImage');

    currentImageElement = imgElement;
    currentFile = fileName;

    cropImage.src = imgElement.src;
    modal.classList.remove('hidden');

    if (cropper) {
        cropper.destroy();
    }

    cropper = new Cropper(cropImage, {
        aspectRatio: CROP_WIDTH / CROP_HEIGHT,
        viewMode: 1,
        dragMode: 'move',
        cropBoxResizable: true,
        cropBoxMovable: true,
        minContainerWidth: 400,
        minContainerHeight: 400,
        imageSmoothingEnabled: true,
        imageSmoothingQuality: 'high',
        background: false,
        modal: false,
        transparent: true
    });

    showSuccessToast('Crop mode enabled');
}

function cancelCrop() {
    const modal = document.getElementById('cropModal');
    modal.classList.add('hidden');
    if (cropper) {
        cropper.destroy();
        cropper = null;
    }
    showErrorToast('Crop cancelled');
}

function saveCrop() {
    if (!cropper) return;

    const canvas = cropper.getCroppedCanvas({
        width: CROP_WIDTH,
        height: CROP_HEIGHT,
        imageSmoothingEnabled: true,
        imageSmoothingQuality: 'high'
    });

    canvas.toBlob((blob) => {
        const croppedFile = new File([blob], currentFile, { type: 'image/png' });
        const previewContainer = document.getElementById('banner-preview');

        // Update preview with cropped image
        const imgPreview = previewContainer.querySelector('img');
        imgPreview.src = URL.createObjectURL(croppedFile);

        cancelCrop();
        showSuccessToast('Image cropped successfully');
    }, 'image/png', 1.0);
}

// Initialize drag and drop on page load
document.addEventListener('DOMContentLoaded', enableDragAndDrop);

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

// Get the form element
const descriptionForm = document.getElementById("descriptionForm");

// Add submit event listener to the form
descriptionForm.addEventListener("submit", async (e) => {
    e.preventDefault();

    try {
        const formData = new FormData(descriptionForm);

        const response = await fetch("/admin/products/main/submit-description", {
            method: "POST",
            body: formData
        });

        if (response.ok) {
            showSuccessToast("Description added successfully!");
            setTimeout(() => {
                window.location.reload();
            }, 500); modal.classList.add("hidden"); // Close modal on success
            descriptionForm.reset(); // Reset form

            // Reset the form to just one pair of inputs
            keyValuePairs.innerHTML = `
                <div class="flex mb-4 space-x-4">
                    <input type="text" name="heading[]" placeholder="Enter heading"
                        class="p-2 border border-gray-300 rounded w-1/3" required>
                    <textarea name="description[]" placeholder="Enter description"
                        class="p-2 border border-gray-300 rounded w-2/3" required></textarea>
                </div>
            `;
        } else {
            showErrorToast("Failed to add description. Please try again.");
        }
    } catch (error) {
        console.error("Error submitting form:", error);
        showErrorToast("An error occurred. Please try again.");
    }
});

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
    showSuccessToast("New description field added");
};

// Optional: Add error handling for required fields
descriptionForm.querySelectorAll('input, textarea').forEach(field => {
    field.addEventListener('invalid', () => {
        showErrorToast("Please fill in all required fields");
    });
});

function deleteDescription(descriptionId, productId) {
    // Split the IDs if they're passed as a comma-separated string
    const ids = descriptionId.split(',');
    const descId = ids[0];
    const prodId = ids[1] || productId;

    fetch(`/admin/products/variant/description/delete/${descId}`, {
        method: 'DELETE',
        headers: {
            'Content-Type': 'application/json'
        }
    })
        .then(response => {
            if (!response.ok) {
                throw new Error('Failed to delete description');
            }
            return response.json();
        })
        .then(() => {
            showSuccessToast('Description deleted successfully');
            // Remove the description item from the DOM
            const descriptionItem = document.querySelector(`.Descriptions-item[data-desc-id="${descId}"]`);
            if (descriptionItem) {
                descriptionItem.remove();
            }
            // Redirect after a short delay to allow toast to be seen
            setTimeout(() => {
                window.location.href = `/admin/products/main/details?product_id=${prodId}`;
            }, 1000);
        })
        .catch(error => {
            console.error('Error:', error);
            showErrorToast('Failed to delete description');
        });
}


// Update/Delete Offer Modal
const updateModal = document.getElementById("updateModal");
const closeUpdateModal = document.getElementById("closeUpdateModal");
const updateOfferForm = document.getElementById("updateOfferForm");
const updateOfferBtn = document.getElementById("updateOfferBtn");
const deleteOfferBtn = document.getElementById("deleteOfferBtn");

// Function to show the modal
function openUpdateModal() {
    updateModal.classList.remove("hidden");
}

// Function to close the modal
function closeModal() {
    updateModal.classList.add("hidden");
}

// Event listener for opening the modal
if (updateOfferBtn) {
    updateOfferBtn.addEventListener("click", openUpdateModal);
}

// Event listener for closing the modal
if (closeUpdateModal) {
    closeUpdateModal.addEventListener("click", closeModal);
}

// Handle form submission
if (updateOfferForm) {
    updateOfferForm.addEventListener("submit", function (event) {
        // Prevent the default form submission behavior
        event.preventDefault();

        // Get form data
        const formData = new FormData(updateOfferForm);

        // Create an object from form data
        const offerData = {
            productId: document.getElementById("productId").value,
            offerId: document.getElementById("offerId").value,
            offerName: document.getElementById("updateOfferName").value,
            offerDetails: document.getElementById("updateOfferDetails").value,
            startDate: document.getElementById("updateStartDate").value,
            endDate: document.getElementById("updateEndDate").value,
            percentage: document.getElementById("updatePercentage").value
        };

        // Send data using fetch API with POST method
        fetch('/admin/products/main/edit/offer', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(offerData)
        })
            .then(response => {
                if (response.ok) {
                    return response.json();
                }
                throw new Error('Network response was not ok');
            })
            .then(data => {
                showSuccessToast('Offer updated successfully!');
                closeModal();

                // Optionally refresh the page or update the UI
                location.reload();
            })
            .catch(error => {
                console.error('Error updating offer:', error);
                showErrorToast('Failed to update offer. Please try again.');
            });
    });
}

 // Wait for the DOM to be fully loaded
 document.addEventListener('DOMContentLoaded', function() {
    // Select all delete buttons
    const deleteButtons = document.querySelectorAll('.delete-btn');
    
    // Add click event listener to each delete button
    deleteButtons.forEach(button => {
        button.addEventListener('click', function() {
            // Get the parent form
            const form = this.closest('.delete-product-form');
            
            // Get the product ID from the hidden input
            const productId = form.querySelector('.product-id').value;
            
            // Confirm before proceeding
            if (confirm(`Are you sure you want to delete product offer`)) {
                // Option 1: Submit the form (traditional approach)
                // form.submit();
                
                // Option 2: Use fetch API (modern approach)
                fetch('/admin/products/main/delete/offer', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        // Include CSRF token if your framework requires it
                        // 'X-CSRF-TOKEN': document.querySelector('meta[name="csrf-token"]').getAttribute('content')
                    },
                    body: JSON.stringify({
                        productId: productId
                    })
                })
                .then(response => {
                    if (response.ok) {
                        return response.json();
                    }
                    throw new Error('Network response was not ok');
                })
                .then(data => {
                    // Handle successful deletion
                    
                    console.log('Product deleted successfully');
                    location.reload();
                    // Remove the product element from the DOM
                    const productElement = form.closest('.border-b');
                    if (productElement) {
                        productElement.remove();
                    }
                })
                .catch(error => {
                    console.error('Error deleting product:', error);
                    showErrorToast('Failed to delete product. Please try again.');
                });
            }
        });
    });
});