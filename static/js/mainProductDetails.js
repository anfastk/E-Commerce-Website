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
            window.toast.success("Description added successfully!");
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
            window.toast.error("Failed to add description. Please try again.");
        }
    } catch (error) {
        console.error("Error submitting form:", error);
        window.toast.error("An error occurred. Please try again.");
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
    window.toast.success("New description field added");
};

// Optional: Add error handling for required fields
descriptionForm.querySelectorAll('input, textarea').forEach(field => {
    field.addEventListener('invalid', () => {
        window.toast.error("Please fill in all required fields");
    });
});


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

        try {
            // Create FormData object
            const formData = new FormData(updateDescriptionsForm);

            // Convert FormData to an object
            const data = {
                description_id: formData.getAll('description_id[]'),
                heading: formData.getAll('heading[]'),
                description: formData.getAll('description[]')
            };

            const response = await fetch(updateDescriptionsForm.action, {
                method: 'PATCH',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(data)
            });

            const responseData = await response.json();

            if (response.ok) {
                window.toast.success('Descriptions updated successfully');
                popupModal.classList.add('hidden');
                // Reload page after a short delay to allow toast to be seen
                setTimeout(() => {
                    window.location.reload();
                }, 1000);
            } else {
                // Show error message from server or fallback
                window.toast.error(responseData.error || 'Failed to update descriptions');
            }
        } catch (error) {
            console.error('Error:', error);
            window.toast.error('An error occurred while updating descriptions');
        }
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
            window.toast.success('Description deleted successfully');
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
            window.toast.error('Failed to delete description');
        });
}


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
        window.toast.error('Please upload an image first');
        return;
    }

    // Show loading toast
    window.toast.success('Uploading image...');

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
                        window.toast.success('Image uploaded successfully');
                        closeUploadPopup();
                        setTimeout(() => {
                            window.location.reload();
                        }, 1000);
                    } else {
                        window.toast.error('Upload failed');
                    }
                })
                .catch(error => {
                    console.error('Upload error:', error);
                    window.toast.error('Failed to upload image');
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
        window.toast.error('Please upload only one image');
        return;
    }

    const file = files[0];
    if (!file.type.startsWith('image/')) {
        window.toast.error('Please upload only image files');
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
        window.toast.success('Image loaded successfully');
    };

    reader.onerror = function () {
        window.toast.error('Failed to load image');
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

    window.toast.success('Crop mode enabled');
}

function cancelCrop() {
    const modal = document.getElementById('cropModal');
    modal.classList.add('hidden');
    if (cropper) {
        cropper.destroy();
        cropper = null;
    }
    window.toast.error('Crop cancelled');
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
        window.toast.success('Image cropped successfully');
    }, 'image/png', 1.0);
}

// Initialize drag and drop on page load
document.addEventListener('DOMContentLoaded', enableDragAndDrop);