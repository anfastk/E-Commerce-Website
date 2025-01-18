const scrollContainer = document.getElementById('product-scroll');
const scrollLeftButton = document.getElementById('scroll-left');
const scrollRightButton = document.getElementById('scroll-right');

scrollLeftButton.addEventListener('click', () => {
    scrollContainer.scrollBy({ left: -300, behavior: 'smooth' });
});

scrollRightButton.addEventListener('click', () => {
    scrollContainer.scrollBy({ left: 300, behavior: 'smooth' });
});

    // Auto-scroll logic (Minimal JavaScript)
    const carouselTrack = document.getElementById('carousel-track');
    const dots = document.querySelectorAll('#dots-container > span');
    let currentIndex = 0;
    const totalSlides = dots.length;

    function updateCarousel() {
      const offset = -currentIndex * 100;
      carouselTrack.style.transform = `translateX(${offset}%)`;
      dots.forEach(dot => dot.classList.remove('bg-gray-700'));
      dots[currentIndex].classList.add('bg-gray-700');
    }

    function autoScroll() {
      currentIndex = (currentIndex + 1) % totalSlides;
      updateCarousel();
    }

    let interval = setInterval(autoScroll, 2000);

    // Manual navigation
    dots.forEach((dot, index) => {
      dot.addEventListener('click', () => {
        currentIndex = index;
        updateCarousel();
        clearInterval(interval); // Reset auto-scroll
        interval = setInterval(autoScroll, 2000);
      });
    });