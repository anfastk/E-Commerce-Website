let cropper = null;
let currentImageElement = null;
let currentFile = null;
const MIN_IMAGES = 1; // Changed from MAX_IMAGES to MIN_IMAGES
const CROP_WIDTH = 400;
const CROP_HEIGHT = 400;
let variantCount = 1;

// Store files for each variant
const variantFiles = new Map();

document.addEventListener('DOMContentLoaded', function () {
  const form = document.querySelector('form');
  if (!form) return;

  // Helper function to show field error
  const showFieldError = (field, message) => {
    field.classList.add('border-red-500');
    // Add error message below the field
    const existingError = field.nextElementSibling?.classList.contains('error-message');
    if (!existingError) {
      const errorDiv = document.createElement('div');
      errorDiv.className = 'error-message text-red-500 text-sm mt-1';
      errorDiv.textContent = message;
      field.parentNode.insertBefore(errorDiv, field.nextSibling);
    }
  };

  // Helper function to clear field error
  const clearFieldError = (field) => {
    field.classList.remove('border-red-500');
    const errorMessage = field.nextElementSibling;
    if (errorMessage?.classList.contains('error-message')) {
      errorMessage.remove();
    }
  };

  form.addEventListener('submit', function (e) {
    e.preventDefault();

    const formData = new FormData(this);
    let isValid = true;
    let errorMessages = [];

    // Clear all previous errors
    form.querySelectorAll('.error-message').forEach(error => error.remove());
    form.querySelectorAll('.border-red-500').forEach(field => field.classList.remove('border-red-500'));

    // Validate required fields
    const requiredFields = this.querySelectorAll('[required]');
    requiredFields.forEach(field => {
      if (!field.value.trim()) {
        isValid = false;
        const fieldName = field.getAttribute('name') || field.getAttribute('id') || 'Field';
        showFieldError(field, `${fieldName} is required`);
        errorMessages.push(`${fieldName} is required`);
      } else {
        clearFieldError(field);
      }
    });

    // Validate numeric fields
    const priceFields = this.querySelectorAll('[id^="regular-price-"], [id^="discounted-price-"], [id^="stock-quantity-"]');
    priceFields.forEach(field => {
      const value = parseFloat(field.value);
      const fieldName = field.getAttribute('name') || field.getAttribute('id') || 'Field';
      
      if (field.value.trim() === '') {
        return; // Skip empty fields, they'll be caught by required validation if needed
      }

      if (isNaN(value)) {
        isValid = false;
        showFieldError(field, `${fieldName} must be a valid number`);
        errorMessages.push(`${fieldName} must be a valid number`);
      } else if (value < 0) {
        isValid = false;
        showFieldError(field, `${fieldName} cannot be negative`);
        errorMessages.push(`${fieldName} cannot be negative`);
      } else {
        clearFieldError(field);
      }
    });

    if (!isValid) {
      // Show toast with all error messages
      window.toast.error('Please fix the following errors:\n' + errorMessages.join('\n'));
      return;
    }

    showLoader();
    fetch(this.action, {
      method: 'POST',
      body: formData
    })
      .then(response => response.json())
      .then(data => {
        hideLoader();
        if (data.code === 200) {
          window.toast.success(data.message || 'Variant added successfully!');
          setTimeout(() => {
            window.location.href = data.redirect || '/admin/products';
          }, 1000);
        } else {
          window.toast.error(data.message || 'Error adding variant');
        }
      })
      .catch(error => {
        hideLoader();
        console.error('Error:', error);
        window.toast.error('Error adding variant. Please try again.');
      });
  });

  // Add real-time validation for numeric fields
  const numericFields = form.querySelectorAll('[id^="regular-price-"], [id^="discounted-price-"], [id^="stock-quantity-"]');
  numericFields.forEach(field => {
    field.addEventListener('input', function() {
      const value = parseFloat(this.value);
      const fieldName = this.getAttribute('name') || this.getAttribute('id') || 'Field';
      
      if (this.value.trim() === '') {
        clearFieldError(this);
        return;
      }

      if (isNaN(value)) {
        showFieldError(this, `${fieldName} must be a valid number`);
      } else if (value < 0) {
        showFieldError(this, `${fieldName} cannot be negative`);
      } else {
        clearFieldError(this);
      }
    });
  });
});

// Image handling functions
function handleFileUpload(input, previewContainerId) {
  const previewContainer = document.getElementById(previewContainerId);
  const files = input.files;
  const variantForm = input.closest('.variant-form');
  const variantIndex = Array.from(document.querySelectorAll('.variant-form')).indexOf(variantForm);

  if (files.length < MIN_IMAGES) {
    window.toast.error(`You must upload at least ${MIN_IMAGES} image`, 'error');
    return;
  }

  // File validation
  for (let file of files) {
    if (!file.type.startsWith('image/')) {
      window.toast.error('Please upload only image files', 'error');
      return;
    }
    if (file.size > 5 * 1024 * 1024) { // 5MB limit
      window.toast.error('Image size should not exceed 5MB', 'error');
      return;
    }
  }

  // Initialize or get the files array for this variant
  if (!variantFiles.has(variantIndex)) {
    variantFiles.set(variantIndex, []);
  }
  const filesArray = variantFiles.get(variantIndex);

  Array.from(files).forEach((file) => {
    const reader = new FileReader();
    reader.onload = function (e) {
      const preview = document.createElement('div');
      preview.className = 'relative border rounded p-2';
      preview.innerHTML = `
        <img src="${e.target.result}" alt="" class="w-full h-40 object-contain" data-original-file="${file.name}" data-variant-index="${variantIndex}" style="background: transparent;">
        <div class="absolute top-2 right-2 flex gap-2">
          <button type="button" class="bg-blue-500 text-white p-1 rounded" onclick="startCrop(this.parentElement.parentElement.querySelector('img'), '${file.name}')">
            Crop
          </button>
          <button type="button" class="bg-red-500 text-white p-1 rounded" onclick="removePreview('${input.id}', '${previewContainerId}', this.parentElement.parentElement, ${variantIndex}, '${file.name}')">
            ×
          </button>
        </div>
      `;
      previewContainer.appendChild(preview);
      filesArray.push(file);
    };
    reader.readAsDataURL(file);
  });

  variantFiles.set(variantIndex, filesArray);
  input.value = '';
}

// Keep your existing functions for cropping, drag-and-drop, etc...
// Just add appropriate toast messages where needed

function removePreview(inputId, previewContainerId, previewElement, variantIndex, fileName) {
  previewElement.remove();
  const filesArray = variantFiles.get(variantIndex);
  const fileIndex = filesArray.findIndex(file => file.name === fileName);
  if (fileIndex > -1) {
    filesArray.splice(fileIndex, 1);
    variantFiles.set(variantIndex, filesArray);
    window.toast.success('Image removed successfully', 'success');
  }
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
    background: false, // Disable the dark background
    modal: false, // Disable the black modal
    transparent: true // Enable transparency
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
    const variantForm = currentImageElement.closest('.variant-form');
    const variantIndex = Array.from(document.querySelectorAll('.variant-form')).indexOf(variantForm);
    const filesArray = variantFiles.get(variantIndex);

    const fileIndex = filesArray.findIndex(file => file.name === currentFile);
    if (fileIndex > -1) {
      filesArray[fileIndex] = croppedFile;
      variantFiles.set(variantIndex, filesArray);
    }

    currentImageElement.src = URL.createObjectURL(croppedFile);
    currentImageElement.style.background = 'transparent';

    cancelCrop();
    window.toast.success('Image cropped successfully', 'success');
  }, 'image/png', 1.0);
}

function enableDragAndDrop(dropAreaId, inputId, previewContainerId) {
  const dropArea = document.getElementById(dropAreaId);
  const input = document.getElementById(inputId);

  if (!dropArea || !input) {
    console.error('Required elements not found');
    return;
  }

  // Enable multiple file selection
  input.setAttribute('multiple', 'multiple');

  ['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => {
    dropArea.addEventListener(eventName, preventDefaults, false);
  });

  function preventDefaults(e) {
    e.preventDefault();
    e.stopPropagation();
  }

  dropArea.addEventListener('dragenter', () => {
    dropArea.classList.add('border-blue-500');
  });

  dropArea.addEventListener('dragleave', () => {
    dropArea.classList.remove('border-blue-500');
  });

  dropArea.addEventListener('drop', (e) => {
    dropArea.classList.remove('border-blue-500');
    const dt = e.dataTransfer;
    const files = dt.files;

    // Create a file input-like object that matches the structure expected by handleFileUpload
    const fileInputObj = {
      files: files,
      id: inputId,
      closest: function (selector) {
        return dropArea.closest(selector);
      }
    };

    handleFileUpload(fileInputObj, previewContainerId);
  });

  input.addEventListener('change', () => {
    handleFileUpload(input, previewContainerId);
  });
}

function addVariant() {
  variantCount++;
  const container = document.getElementById('variants-container');

  // Get the image source from the first variant if it exists
  const firstVariantImage = document.querySelector('.variant-form img');
  const imageSource = firstVariantImage ? firstVariantImage.src : '';

  const newVariant = document.createElement('div');
  newVariant.classList.add('variant-form', 'mb-8');

  newVariant.innerHTML = `
        <div class="grid grid-cols-2 gap-8">
            <div>
                <label for="product-name-${variantCount}" class="block text-sm font-medium text-gray-700">Product Name</label>
                <input type="text" id="product-name-${variantCount}" name="product-name[]" placeholder="Product Name"
                       class="mt-1 block w-full border border-gray-300 rounded-md p-2" required>

                <label for="product-summary-${variantCount}" class="block mt-4 text-sm font-medium text-gray-700">Product Summary</label>
                <input type="text" id="product-summary-${variantCount}" name="product-summary[]" placeholder="Product Summary"
                       class="mt-1 block w-full border border-gray-300 rounded-md p-2" required>

                <label for="size-${variantCount}" class="block mt-4 text-sm font-medium text-gray-700">Size</label>
                <input type="text" id="size-${variantCount}" name="size[]" placeholder="Product Size"
                       class="mt-1 block w-full border border-gray-300 rounded-md p-2">

                <label for="color-${variantCount}" class="block mt-4 text-sm font-medium text-gray-700">Colour</label>
                <input type="text" id="color-${variantCount}" name="color[]" placeholder="Product Colour"
                       class="mt-1 block w-full border border-gray-300 rounded-md p-2">

                <label for="ram-${variantCount}" class="block mt-4 text-sm font-medium text-gray-700">Ram</label>
                <input type="text" id="ram-${variantCount}" name="ram[]" placeholder="Product Ram"
                       class="mt-1 block w-full border border-gray-300 rounded-md p-2">

                <label for="storage-${variantCount}" class="block mt-4 text-sm font-medium text-gray-700">Storage</label>
                <input type="text" id="storage-${variantCount}" name="storage[]" placeholder="Product Storage"
                       class="mt-1 block w-full border border-gray-300 rounded-md p-2">
            </div>

            <div class="flex justify-center">
                <div class="relative">
                    <img src="${imageSource}" alt="Product Image" class="object-contain">
                </div>
            </div>
        </div>
        <div class="grid grid-cols-2 gap-8 mt-6">
            <div>
                <label for="regular-price-${variantCount}" class="block text-sm font-medium text-gray-700">Regular Price</label>
                <input type="text" id="regular-price-${variantCount}" name="regular-price[]" placeholder="$110.40"
                       class="mt-1 block w-full border border-gray-300 rounded-md p-2">
            </div>
            <div>
                <label for="sale-price-${variantCount}" class="block text-sm font-medium text-gray-700">Sale Price</label>
                <input type="text" id="sale-price-${variantCount}" name="sale-price[]" placeholder="$450"
                       class="mt-1 block w-full border border-gray-300 rounded-md p-2">
            </div>
        </div>

        <div class="grid grid-cols-2 gap-8 mt-6">
            <div>
                <label for="stock-quantity-${variantCount}" class="block text-sm font-medium text-gray-700">Stock Quantity</label>
                <input type="number" id="stock-quantity-${variantCount}" name="stock-quantity[]" placeholder="21"
                       class="mt-1 block w-full border border-gray-300 rounded-md p-2">
            </div>
            <div>
                <label for="sku-${variantCount}" class="block text-sm font-medium text-gray-700">SKU</label>
                <input type="text" id="sku-${variantCount}" name="sku[]" placeholder="21"
                       class="mt-1 block w-full border border-gray-300 rounded-md p-2">
            </div>
        </div>

        <div class="mb-6">
            <label class="block text-black font-medium mb-2">Product Images</label>
            <div id="banner-drop-area-${variantCount}"
                class="border-dashed border-2 rounded px-4 py-6 text-center bg-white">
                <input id="banner-input-${variantCount}" type="file" accept="image/*" multiple class="hidden"
                    name="product_images[]" multiple/>
                <p class="text-gray-500">
                    Drop your images here, or
                    <span class="text-blue-500 underline cursor-pointer"
                        onclick="document.getElementById('banner-input-${variantCount}').click()">browse</span>
                </p>
                <p class="text-sm text-gray-500 mt-2">Upload at least ${MIN_IMAGES} image</p>
            </div>
            <div id="banner-preview-${variantCount}" class="mt-4 grid grid-cols-3 gap-4"></div>
        </div>
  `;

  container.appendChild(newVariant);

  // Enable drag-and-drop for the new variant
  enableDragAndDrop(`banner-drop-area-${variantCount}`, `banner-input-${variantCount}`, `banner-preview-${variantCount}`);
}

// Initialize on page load
document.addEventListener('DOMContentLoaded', () => {
  enableDragAndDrop('banner-drop-area', 'banner-input', 'banner-preview');
});