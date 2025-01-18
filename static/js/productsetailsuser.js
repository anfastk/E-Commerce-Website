
// Show modal
document.getElementById('showCouponButton').addEventListener('click', () => {
    document.getElementById('couponModal').classList.remove('hidden');
});

// Hide modal
document.getElementById('closeModalButton').addEventListener('click', () => {
    document.getElementById('couponModal').classList.add('hidden');
});

// Copy coupon code to clipboard
function copyToClipboard(couponCode) {
    navigator.clipboard.writeText(couponCode).then(() => {
        alert(`Coupon code "${couponCode}" copied to clipboard!`);
    });
}

document.addEventListener('DOMContentLoaded', () => {
    const minusButton = document.querySelector('#minus-button');
    const plusButton = document.querySelector('#plus-button');
    const quantityInput = document.querySelector('#quantity-input');

    minusButton.addEventListener('click', () => {
        let currentValue = parseInt(quantityInput.value);
        if (currentValue > 1) {
            quantityInput.value = currentValue - 1;
        }
    });

    plusButton.addEventListener('click', () => {
        let currentValue = parseInt(quantityInput.value);
        quantityInput.value = currentValue + 1;
    });
});



function changeImage(thumbnail) {
    const mainImage = document.getElementById('main-image');
    const thumbnailContainer = document.getElementById('thumbnail-container');
    
    // Swap the main image with the clicked thumbnail
    const tempSrc = mainImage.src;
    mainImage.src = thumbnail.src;
    thumbnail.src = tempSrc;

    // Scroll the clicked thumbnail into view
    thumbnail.scrollIntoView({ behavior: 'smooth', block: 'nearest', inline: 'center' });
}

function scrollThumbnails(direction) {
    const container = document.getElementById('thumbnail-container');
    const scrollAmount = 120; // Adjust scroll amount as needed
    container.scrollBy({ left: direction * scrollAmount, behavior: 'smooth' });
}