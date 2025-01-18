function handleFileUpload(input, previewContainerId) {
  const previewContainer = document.getElementById(previewContainerId);

  // Clear previous previews
  previewContainer.innerHTML = '';

  const files = input.files; // Get all files

  if (files.length > 0) {
    Array.from(files).forEach((file, index) => {
      const reader = new FileReader();

      reader.onload = function (e) {
        const preview = document.createElement('div');
        preview.className = 'flex items-center border rounded px-4 py-2 mb-2';
        preview.innerHTML = `
          <img src="${e.target.result}" alt="" class="w-12 h-12 mr-4">
          <p class="flex-1 text-gray-700">${file.name}</p>
          <button type="button" class="text-red-500" onclick="removePreview('${input.id}', '${previewContainerId}', ${index})">&times;</button>
        `;
        previewContainer.appendChild(preview);
      };

      reader.readAsDataURL(file);
    });
  }
}

function removePreview(inputId, previewContainerId, fileIndex) {
  const input = document.getElementById(inputId);
  const previewContainer = document.getElementById(previewContainerId);

  const files = Array.from(input.files);
  files.splice(fileIndex, 1);

  // Update file input
  const dataTransfer = new DataTransfer();
  files.forEach((file) => dataTransfer.items.add(file));
  input.files = dataTransfer.files;

  // Refresh previews
  handleFileUpload(input, previewContainerId);
}

function enableDragAndDrop(dropAreaId, inputId, previewContainerId) {
  const dropArea = document.getElementById(dropAreaId);
  const input = document.getElementById(inputId);

  dropArea.addEventListener('dragover', (e) => {
    e.preventDefault();
    dropArea.classList.add('border-blue-500');
  });

  dropArea.addEventListener('dragleave', () => {
    dropArea.classList.remove('border-blue-500');
  });

  dropArea.addEventListener('drop', (e) => {
    e.preventDefault();
    dropArea.classList.remove('border-blue-500');

    const files = e.dataTransfer.files;
    const dataTransfer = new DataTransfer();

    // Add all dropped files
    Array.from(files).forEach((file) => dataTransfer.items.add(file));

    input.files = dataTransfer.files;
    handleFileUpload(input, previewContainerId);
  });

  input.addEventListener('change', () => {
    handleFileUpload(input, previewContainerId);
  });
}

document.addEventListener('DOMContentLoaded', () => {
  enableDragAndDrop('banner-drop-area', 'banner-input', 'banner-preview');
});

let variantCount = 1;

function addVariant() {
  variantCount++;
  const container = document.getElementById('variants-container');

  const newVariant = document.createElement('div');
  newVariant.classList.add('variant-form', 'mb-8');

  newVariant.innerHTML = `
    <div class="grid grid-cols-2 gap-8">
              <div>
                  <label for="product-name-${variantCount}" class="block text-sm font-medium text-gray-700">Product Name</label>
                  <input type="text" id="product-name-${variantCount}" name="product-name[]" placeholder="Product Name"
                         class="mt-1 block w-full border border-gray-300 rounded-md p-2" required>

                  <label for="product-summary-${variantCount}" class="block mt-4 text-sm font-medium text-gray-700">Product Summary</label>
                  <input type="text" id="product-summary-${variantCount}" name="product-summary[]" placeholder="Product Summery"
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
                  <!-- Placeholder for Image (Not Image Upload) -->
                  <div class="relative">
                      <img src="{{ .Images.ProductImages }}" alt="Product Image" class="object-contain">
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

          <div>
        <label class="block text-sm font-medium text-gray-700">Product Images</label>
        <div id="banner-drop-area-${variantCount}" class="border-dashed border-2 rounded px-4 py-6 text-center bg-white">
          <input id="banner-input-${variantCount}" type="file" accept="image/*" multiple class="hidden" name="product_images[]" />
          <p class="text-gray-500">
            Drop your images here, or
            <span class="text-blue-500 underline cursor-pointer" onclick="document.getElementById('banner-input-${variantCount}').click()">browse</span>
          </p>
        </div>
        <div id="banner-preview-${variantCount}" class="mt-4 space-y-2"></div>
      </div>
    </div>
      `;

  container.appendChild(newVariant);

  // Enable drag-and-drop for the new variant
  enableDragAndDrop(`banner-drop-area-${variantCount}`, `banner-input-${variantCount}`, `banner-preview-${variantCount}`);
}