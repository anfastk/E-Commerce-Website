let cropper = null;
let currentImageElement = null;
let currentFile = null;
const MAX_IMAGES = 2;
const CROP_WIDTH = 400;
const CROP_HEIGHT = 400;
let variantCount = 1;

// Store files for each variant
const variantFiles = new Map();

function handleFileUpload(input, previewContainerId) {
  const previewContainer = document.getElementById(previewContainerId);
  const files = input.files;
  const variantForm = input.closest('.variant-form');
  const variantIndex = Array.from(document.querySelectorAll('.variant-form')).indexOf(variantForm);

  if (previewContainer.children.length + files.length > MAX_IMAGES) {
    alert(`You can only upload a maximum of ${MAX_IMAGES} images.`);
    return;
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


function removePreview(inputId, previewContainerId, previewElement, variantIndex, fileName) {
  previewElement.remove();
  const filesArray = variantFiles.get(variantIndex);
  const fileIndex = filesArray.findIndex(file => file.name === fileName);
  if (fileIndex > -1) {
    filesArray.splice(fileIndex, 1);
    variantFiles.set(variantIndex, filesArray);
  }
}

// Add this function to handle form submission
document.querySelector('form').addEventListener('submit', function (e) {
  e.preventDefault();

  const formData = new FormData(this);
  
  // Add variant index to identify which images belong to which variant
  variantFiles.forEach((files, variantIndex) => {
    files.forEach(file => {
      // Append files with variant index in the name
      formData.append(`product_images[${variantIndex}][]`, file);
    });
  });

  // Submit the form with fetch
  fetch(this.action, {
    method: 'POST',
    body: formData
  })
    .then(response => response.json())
    .then(data => {
      if (data.code === 200) {
        window.location.href = data.redirect || '/admin/products';
      } else {
        alert(data.message || 'Error uploading files');
      }
    })
    .catch(error => {
      console.error('Error:', error);
      alert('Error uploading files');
    });
});

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

  // Convert canvas to blob, maintaining transparency
  canvas.toBlob((blob) => {
    // Create a new file from the blob
    const croppedFile = new File([blob], currentFile, { type: 'image/png' }); // Use PNG to preserve transparency

    // Update the variantFiles map to replace the original file with the cropped file
    const variantForm = currentImageElement.closest('.variant-form');
    const variantIndex = Array.from(document.querySelectorAll('.variant-form')).indexOf(variantForm);
    const filesArray = variantFiles.get(variantIndex);

    // Find and replace the original file
    const fileIndex = filesArray.findIndex(file => file.name === currentFile);
    if (fileIndex > -1) {
      filesArray[fileIndex] = croppedFile;
      variantFiles.set(variantIndex, filesArray);
    }

    // Update the preview image
    currentImageElement.src = URL.createObjectURL(croppedFile);
    currentImageElement.style.background = 'transparent';

    // Close the crop modal
    cancelCrop();
  }, 'image/png', 1.0); // Use PNG format with maximum quality
}


// Rest of your existing code remains the same...

function removePreview(inputId, previewContainerId, previewElement) {
  previewElement.remove();
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
          closest: function(selector) {
              return dropArea.closest(selector);
          }
      };
      
      handleFileUpload(fileInputObj, previewContainerId);
  });
  
  input.addEventListener('change', () => {
      handleFileUpload(input, previewContainerId);
  });
}

// Modified handleFileUpload function to handle both drag-drop and file input
function handleFileUpload(input, previewContainerId) {
  const previewContainer = document.getElementById(previewContainerId);
  if (!previewContainer) {
      console.error('Preview container not found:', previewContainerId);
      return;
  }
  
  const files = input.files;
  const variantForm = input.closest('.variant-form');
  const variantIndex = variantForm ? 
      Array.from(document.querySelectorAll('.variant-form')).indexOf(variantForm) : 
      0;
  
  if (previewContainer.children.length + files.length > MAX_IMAGES) {
      alert(`You can only upload a maximum of ${MAX_IMAGES} images.`);
      return;
  }

  // Initialize or get the files array for this variant
  if (!variantFiles.has(variantIndex)) {
      variantFiles.set(variantIndex, []);
  }
  const filesArray = variantFiles.get(variantIndex);

  Array.from(files).forEach((file) => {
      if (!file.type.startsWith('image/')) {
          alert('Please upload only image files.');
          return;
      }

      const reader = new FileReader();
      reader.onload = function(e) {
          const preview = document.createElement('div');
          preview.className = 'relative border rounded p-2';
          preview.innerHTML = `
              <img src="${e.target.result}" alt="" class="w-full h-40 object-cover" data-original-file="${file.name}">
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
  if (input.value) input.value = '';
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
                <p class="text-sm text-gray-500 mt-2">Upload up to 6 images</p>
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

