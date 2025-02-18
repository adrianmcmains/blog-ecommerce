// Intersection Observer for animations
const observerOptions = {
    root: null,
    rootMargin: '0px',
    threshold: 0.1
};

const observer = new IntersectionObserver((entries, observer) => {
    entries.forEach(entry => {
        if (entry.isIntersecting) {
            entry.target.classList.add('animate');
            observer.unobserve(entry.target);
        }
    });
}, observerOptions);

// Animate elements when they come into view
document.addEventListener('DOMContentLoaded', () => {
    const animatedElements = document.querySelectorAll('.animate-on-scroll');
    animatedElements.forEach(el => observer.observe(el));

    // Initialize counters
    initCounters();
    
    // Initialize product carousel
    initProductCarousel();
    
    // Initialize smooth scroll
    initSmoothScroll();
    
    // Initialize floating cart
    initFloatingCart();
});

// Counter animation
function initCounters() {
    const counters = document.querySelectorAll('.counter');
    
    counters.forEach(counter => {
        const target = +counter.getAttribute('data-target');
        const duration = 2000; // 2 seconds
        const start = 0;
        const increment = target / (duration / 16); // 60 FPS
        
        let current = start;
        
        const updateCounter = () => {
            current += increment;
            if (current < target) {
                counter.textContent = Math.round(current);
                requestAnimationFrame(updateCounter);
            } else {
                counter.textContent = target;
            }
        };
        
        observer.observe(counter);
        counter.addEventListener('animate', updateCounter);
    });
}

// Product carousel
function initProductCarousel() {
    const carousel = document.querySelector('.product-carousel');
    if (!carousel) return;

    let isDown = false;
    let startX;
    let scrollLeft;

    carousel.addEventListener('mousedown', (e) => {
        isDown = true;
        carousel.classList.add('active');
        startX = e.pageX - carousel.offsetLeft;
        scrollLeft = carousel.scrollLeft;
    });

    carousel.addEventListener('mouseleave', () => {
        isDown = false;
        carousel.classList.remove('active');
    });

    carousel.addEventListener('mouseup', () => {
        isDown = false;
        carousel.classList.remove('active');
    });

    carousel.addEventListener('mousemove', (e) => {
        if (!isDown) return;
        e.preventDefault();
        const x = e.pageX - carousel.offsetLeft;
        const walk = (x - startX) * 2;
        carousel.scrollLeft = scrollLeft - walk;
    });
}

// Smooth scroll
function initSmoothScroll() {
    document.querySelectorAll('a[href^="#"]').forEach(anchor => {
        anchor.addEventListener('click', function (e) {
            e.preventDefault();
            const target = document.querySelector(this.getAttribute('href'));
            if (target) {
                target.scrollIntoView({
                    behavior: 'smooth',
                    block: 'start'
                });
            }
        });
    });
}

// Floating cart
function initFloatingCart() {
    const cart = document.querySelector('.floating-cart');
    if (!cart) return;

    let cartItems = [];
    const cartCount = cart.querySelector('.cart-count');
    const cartTotal = cart.querySelector('.cart-total');

    // Add to cart function
    window.addToCart = function(productId, name, price) {
        cartItems.push({ id: productId, name, price });
        updateCart();
        
        // Show notification
        showNotification(`Added ${name} to cart`);
    };

    function updateCart() {
        cartCount.textContent = cartItems.length;
        cartTotal.textContent = `$${cartItems.reduce((sum, item) => sum + item.price, 0).toFixed(2)}`;
        
        // Animate cart icon
        cart.classList.add('bounce');
        setTimeout(() => cart.classList.remove('bounce'), 300);
    }

    function showNotification(message) {
        const notification = document.createElement('div');
        notification.className = 'cart-notification';
        notification.textContent = message;
        document.body.appendChild(notification);
        
        setTimeout(() => {
            notification.classList.add('show');
            setTimeout(() => {
                notification.classList.remove('show');
                setTimeout(() => notification.remove(), 300);
            }, 2000);
        }, 100);
    }
}

// Back to top button
window.onscroll = function() {
    const backToTop = document.querySelector('.back-to-top');
    if (!backToTop) return;
    
    if (document.body.scrollTop > 500 || document.documentElement.scrollTop > 500) {
        backToTop.classList.add('show');
    } else {
        backToTop.classList.remove('show');
    }
};