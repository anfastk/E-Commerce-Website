const track = document.getElementById('carousel-track');
const slides = track.children;
let currentIndex = 0;

function scrollCarousel() {
    currentIndex = (currentIndex + 1) % slides.length;
    track.style.transform = `translateX(-${currentIndex * 100}%)`;
}

setInterval(scrollCarousel, 3000);