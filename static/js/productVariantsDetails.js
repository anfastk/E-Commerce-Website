// Modal Functionality for Specifications
const openPopup = document.getElementById("openPopup");
const popupModal = document.getElementById("popupModal");
const closePopup = document.getElementById("closePopup");
const specificationsForm = document.getElementById("specificationsForm");

openPopup.addEventListener("click", () => {
    popupModal.classList.remove("hidden");
});

closePopup.addEventListener("click", () => {
    popupModal.classList.add("hidden");
});
 
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

specificationsForm.addEventListener("submit", function (event) {
    event.preventDefault();
    const formData = new FormData(this);
    fetch(this.action, {
        method: 'POST',
        body: formData
    })
        .then(response => {
            if (!response.ok) throw new Error('Network response was not ok');
            return response.json();
        })
        .then(data => {
            showSuccessToast('Specifications saved successfully!');
            setTimeout(() => {
                specificationsForm.reset();
                popupModal.classList.add("hidden");
                window.location.reload();
            }, 1500);
        })
        .catch(error => {
            console.error('Error:', error);
            showErrorToast('Failed to save specifications. Please try again.');
        });
});

// Update/Delete Specification Modal
document.addEventListener('DOMContentLoaded', () => {
    const openPopupBtn = document.getElementById('openUpdatePopup');
    const closePopupBtn = document.getElementById('closeUpdatePopup');
    const popupModal = document.getElementById('updatePopupModal');
    const updateSpecificationsForm = document.getElementById('updateSpecificationsForm');

    openPopupBtn.addEventListener('click', () => {
        popupModal.classList.remove('hidden');
    });

    closePopupBtn.addEventListener('click', () => {
        popupModal.classList.add('hidden');
    });

    popupModal.addEventListener('click', (event) => {
        if (event.target === popupModal) {
            popupModal.classList.add('hidden');
        }
    });

    updateSpecificationsForm.addEventListener('submit', async (event) => {
        event.preventDefault();
        const formData = new FormData(updateSpecificationsForm);
        const data = {
            specification_id: formData.getAll('specification_id[]'),
            specification_key: formData.getAll('specification_key[]'),
            specification: formData.getAll('specification[]')
        };

        try {
            const response = await fetch(updateSpecificationsForm.action, {
                method: 'PATCH',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(data)
            });
            const responseData = await response.json();
            if (response.ok) {
                showSuccessToast('Specifications updated successfully!');
                setTimeout(() => {
                    popupModal.classList.add('hidden');
                    window.location.reload();
                }, 1500);
            } else {
                showErrorToast(responseData.error || 'Failed to update specifications');
            }
        } catch (error) {
            console.error('Error:', error);
            showErrorToast('An error occurred while updating specifications');
        }
    });
});

function deleteSpecification(specificationId, productId) {
    const ids = specificationId.split(',');
    const descId = ids[0];
    const prodId = ids[1] || productId;

    fetch(`/admin/products/variant/specification/delete/${descId}`, {
        method: 'DELETE',
        headers: { 'Content-Type': 'application/json' }
    })
        .then(response => {
            if (response.ok) {
                showSuccessToast('Specification deleted successfully!');
                setTimeout(() => {
                    window.location.href = `/admin/products/variant/detail?variant_id=${prodId}`;
                }, 1500);
            } else {
                showErrorToast('Failed to delete specification');
            }
        })
        .catch(error => {
            console.error('Error:', error);
            showErrorToast('An error occurred while deleting specification');
        });
}

// Image Handling
document.addEventListener('DOMContentLoaded', function () {
    document.querySelectorAll('.options-btn').forEach((btn) => {
        btn.addEventListener('click', function () {
            document.querySelectorAll('.options').forEach(dropdown => {
                if (dropdown !== this.nextElementSibling) {
                    dropdown.classList.add('hidden');
                }
            });
            const dropdown = this.nextElementSibling;
            dropdown.classList.toggle('hidden');
        });
    });

    document.addEventListener('click', function (event) {
        const optionsButtons = document.querySelectorAll('.options-btn');
        let clickedOnOptionsButton = false;
        optionsButtons.forEach(btn => {
            if (btn.contains(event.target)) {
                clickedOnOptionsButton = true;
            }
        });
        if (!clickedOnOptionsButton) {
            document.querySelectorAll('.options').forEach(menu => {
                menu.classList.add('hidden');
            });
        }
    });

    document.querySelectorAll('#openUploadPopup').forEach(button => {
        button.addEventListener('click', function () {
            const imageId = this.getAttribute('data-image-id');
            document.getElementById('current-image-id').value = imageId;
            document.getElementById('imageUploadPopup').classList.remove('hidden');
        });
    });

    let cropper = null;
    let currentImageElement = null;
    let currentFile = null;
    const CROP_WIDTH = 400;
    const CROP_HEIGHT = 400;

    function closeUploadPopup() {
        document.getElementById('imageUploadPopup').classList.add('hidden');
        document.getElementById('banner-preview').innerHTML = '';
        document.getElementById('current-image-id').value = '';
    }

    function confirmUpload() {
        const previewImg = document.getElementById('banner-preview').querySelector('img');
        const imageId = document.getElementById('current-image-id').value;

        if (!previewImg) {
            showErrorToast('Please upload an image first');
            return;
        }

        if (!imageId) {
            showErrorToast('No image selected for replacement');
            return;
        }

        fetch(previewImg.src)
            .then(res => res.blob())
            .then(blob => {
                const formData = new FormData();
                formData.append('product_image', blob, 'uploaded-image.png');
                formData.append('image_id', imageId);

                fetch('/admin/products/variant/image/change', {
                    method: 'POST',
                    body: formData
                })
                    .then(response => response.json())
                    .then(data => {
                        if (data.status === "Success") {
                            showSuccessToast('Image replaced successfully');
                            setTimeout(() => {
                                window.location.reload();
                            }, 1500);
                        } else {
                            showErrorToast(data.message || 'Upload failed. Please try again.');
                        }
                    })
                    .catch(error => {
                        console.error('Upload error:', error);
                        showErrorToast('Upload failed. Please check your connection and try again.');
                    });
            });
    }

    function handleFileUpload(files) {
        const previewContainer = document.getElementById('banner-preview');
        previewContainer.innerHTML = '';

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
            showSuccessToast('Image loaded successfully.');
        };
        reader.onerror = function () {
            showErrorToast('Error reading file. Please try again.');
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
            const imgPreview = previewContainer.querySelector('img');
            imgPreview.src = URL.createObjectURL(croppedFile);

            cancelCrop();
            showSuccessToast('Image cropped successfully');
        }, 'image/png', 1.0);
    }

    function enableDragAndDrop() {
        const dropArea = document.getElementById('banner-drop-area');
        const input = document.getElementById('banner-input');

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

    function cancelCrop() {
        const modal = document.getElementById('cropModal');
        modal.classList.add('hidden');
        if (cropper) {
            cropper.destroy();
            cropper = null;
        }
    }

    window.closeUploadPopup = closeUploadPopup;
    window.confirmUpload = confirmUpload;
    window.startCrop = startCrop;
    window.cancelCrop = cancelCrop;
    window.saveCrop = saveCrop;

    enableDragAndDrop();
});

function handleSubmit(event) {
    event.preventDefault();
    const form = event.target;
    const isRecovery = form.querySelector('button').textContent.trim() === 'RECOVER PRODUCT';

    fetch(form.action, {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
    })
        .then(response => {
            if (!response.ok) throw new Error('Network response was not ok');
            if (isRecovery) {
                showSuccessToast('Product recovered successfully!');
            } else {
                showSuccessToast('Product deleted successfully!');
            }
            setTimeout(() => {
                window.location.reload();
            }, 500);
        })
        .catch(error => {
            console.error('Error:', error);
            if (isRecovery) {
                showErrorToast('Failed to recover product. Please try again.');
            } else {
                showErrorToast('Failed to delete product. Please try again.');
            }
        });
}