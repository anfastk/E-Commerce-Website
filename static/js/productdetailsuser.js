const thumbnailWrappers = document.querySelectorAll('.thumbnail-wrapper');

function changeImage(thumbnailImg) {
    const mainImage = document.getElementById('main-image');
    mainImage.src = thumbnailImg.src;

    thumbnailWrappers.forEach(wrapper => {
        wrapper.classList.remove('active');
    });

    thumbnailImg.parentElement.classList.add('active');
}

function scrollThumbnails(direction) {
    const container = document.getElementById('thumbnail-container');
    const scrollAmount = 160;
    container.scrollBy({
        left: direction * scrollAmount,
        behavior: 'smooth'
    });
}

thumbnailWrappers[0].classList.add('active');


const mainImage = document.getElementById('main-image');
const zoomContainer = document.querySelector('.zoom-container');
const zoomImage = document.getElementById('zoom-image');
const mainImageContainer = document.querySelector('.main-image-container');
const zoomHint = document.querySelector('.zoom-hint');
const magnifierBox = document.querySelector('.magnifier-box');

const ZOOM_LEVEL = 2.5;

function changeImage(thumbnailImg) {
    mainImage.src = thumbnailImg.src;
    zoomImage.src = thumbnailImg.src;

    thumbnailWrappers.forEach(wrapper => {
        wrapper.classList.remove('active');
    });

    thumbnailImg.parentElement.classList.add('active');
}

function scrollThumbnails(direction) {
    const container = document.getElementById('thumbnail-container');
    const scrollAmount = 160;
    container.scrollBy({
        left: direction * scrollAmount,
        behavior: 'smooth'
    });
}

function initZoom() {
    zoomImage.style.width = (mainImage.offsetWidth * ZOOM_LEVEL) + 'px';
    zoomImage.style.height = (mainImage.offsetHeight * ZOOM_LEVEL) + 'px';
}

function handleZoom(e) {
    const mainRect = mainImage.getBoundingClientRect();
    const magnifierRect = magnifierBox.getBoundingClientRect();
    const zoomRect = zoomContainer.getBoundingClientRect();

    // Calculate mouse position relative to the image
    const mouseX = e.clientX - mainRect.left;
    const mouseY = e.clientY - mainRect.top;

    // Calculate the proportional position (0 to 1)
    const proportionalX = Math.max(0, Math.min(1, mouseX / mainRect.width));
    const proportionalY = Math.max(0, Math.min(1, mouseY / mainRect.height));

    // Position magnifier box
    const magnifierX = Math.max(0, Math.min(
        mainRect.width - magnifierBox.offsetWidth,
        mouseX - magnifierBox.offsetWidth / 2
    ));
    const magnifierY = Math.max(0, Math.min(
        mainRect.height - magnifierBox.offsetHeight,
        mouseY - magnifierBox.offsetHeight / 2
    ));

    magnifierBox.style.left = `${magnifierX}px`;
    magnifierBox.style.top = `${magnifierY}px`;

    // Calculate zoom position
    const zoomX = (zoomImage.offsetWidth - zoomRect.width) * proportionalX;
    const zoomY = (zoomImage.offsetHeight - zoomRect.height) * proportionalY;

    zoomImage.style.transform = `translate(-${zoomX}px, -${zoomY}px)`;
}

// Event listeners
mainImageContainer.addEventListener('mouseenter', function (e) {
    zoomContainer.style.display = 'block';
    magnifierBox.style.display = 'block';
    zoomHint.style.display = 'none';
    zoomImage.src = mainImage.src;
    initZoom();
    handleZoom(e); // Initialize zoom position
});

mainImageContainer.addEventListener('mouseleave', function () {
    zoomContainer.style.display = 'none';
    magnifierBox.style.display = 'none';
    zoomHint.style.display = 'block';
});

mainImageContainer.addEventListener('mousemove', handleZoom);

// Show zoom hint on page load
zoomHint.style.display = 'block';

// Initialize first thumbnail as active and set initial zoom image
thumbnailWrappers[0].classList.add('active');
zoomImage.src = mainImage.src;

// Handle window resize
window.addEventListener('resize', initZoom);

// Initialize zoom when image loads
mainImage.addEventListener('load', initZoom);

function showTab(tab) {
    document.querySelectorAll('.tab-content').forEach(content => content.classList.add('hidden'));
    document.getElementById(`content-${tab}`).classList.remove('hidden');
    document.querySelectorAll('button').forEach(tab => tab.classList.remove('tab-active'));
    document.getElementById(`tab-${tab}`).classList.add('tab-active');
}