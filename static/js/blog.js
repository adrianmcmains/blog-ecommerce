// Blog specific JavaScript
console.log('Blog scripts loaded');// Initialize blog functionality
document.addEventListener('DOMContentLoaded', function() {
    initBlog();
});

function initBlog() {
    initializeTableOfContents();
    initializeReadingProgress();
    initializeSocialShare();
    initializeComments();
    initializeNewsletterForm();
    initializeAnimations();
    initializeFeaturedSlider();
}

// Table of Contents
function initializeTableOfContents() {
    const toc = document.querySelector('.toc-wrapper');
    if (!toc) return;

    // Add active state to current section
    const observerOptions = {
        root: null,
        rootMargin: '0px',
        threshold: 0.5
    };

    const headings = document.querySelectorAll('h2, h3');
    const observerCallback = (entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                const id = entry.target.getAttribute('id');
                document.querySelectorAll('.toc-content a').forEach(link => {
                    link.classList.remove('active');
                    if (link.getAttribute('href') === `#${id}`) {
                        link.classList.add('active');
                    }
                });
            }
        });
    };

    const observer = new IntersectionObserver(observerCallback, observerOptions);
    headings.forEach(heading => observer.observe(heading));

    // Toggle TOC on mobile
    const toggleButtons = document.querySelectorAll('.toggle-toc, .toggle-toc-mobile');
    toggleButtons.forEach(button => {
        button.addEventListener('click', function() {
            const content = this.closest('.toc-wrapper, .toc-mobile').querySelector('.toc-content');
            content.style.maxHeight = content.style.maxHeight ? null : content.scrollHeight + 'px';
            this.classList.toggle('active');
        });
    });
}

// Reading Progress
function initializeReadingProgress() {
    const progressBar = document.querySelector('.reading-progress');
    if (!progressBar) return;

    window.addEventListener('scroll', function() {
        const winScroll = document.body.scrollTop || document.documentElement.scrollTop;
        const height = document.documentElement.scrollHeight - document.documentElement.clientHeight;
        const scrolled = (winScroll / height) * 100;
        progressBar.style.width = scrolled + '%';
    });
}

// Social Share
function initializeSocialShare() {
    document.querySelectorAll('.share-btn').forEach(button => {
        button.addEventListener('click', function() {
            const type = this.dataset.share;
            const url = this.dataset.url;
            const title = this.dataset.title;

            let shareUrl;
            switch(type) {
                case 'twitter':
                    shareUrl = `https://twitter.com/intent/tweet?url=${url}&text=${title}`;
                    break;
                case 'facebook':
                    shareUrl = `https://www.facebook.com/sharer/sharer.php?u=${url}`;
                    break;
                case 'linkedin':
                    shareUrl = `https://www.linkedin.com/shareArticle?mini=true&url=${url}&title=${title}`;
                    break;
            }

            window.open(shareUrl, '_blank', 'width=600,height=400');
        });
    });
}

// Comments System
function initializeComments() {
    const commentForm = document.getElementById('comment-form');
    if (!commentForm) return;

    commentForm.addEventListener('submit', async function(e) {
        e.preventDefault();

        const comment = {
            name: this.querySelector('#name').value,
            email: this.querySelector('#email').value,
            comment: this.querySelector('#comment').value,
            postId: window.location.pathname
        };

        try {
            const response = await fetch('/api/comments', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(comment)
            });

            if (response.ok) {
                showNotification('Comment submitted successfully!');
                this.reset();
                loadComments();
            } else {
                throw new Error('Failed to submit comment');
            }
        } catch (error) {
            showNotification('Error submitting comment', 'error');
        }
    });

    loadComments();
}

async function loadComments() {
    const container = document.getElementById('comments-container');
    if (!container) return;

    try {
        const response = await fetch(`/api/comments?postId=${window.location.pathname}`);
        const comments = await response.json();

        container.innerHTML = comments.map(comment => `
            <div class="comment">
                <div class="comment-header">
                    <span class="comment-author">${comment.name}</span>
                    <span class="comment-date">${formatDate(comment.date)}</span>
                </div>
                <div class="comment-content">
                    ${comment.comment}
                </div>
            </div>
        `).join('');
    } catch (error) {
        console.error('Error loading comments:', error);
    }
}

// Newsletter Form
function initializeNewsletterForm() {
    const form = document.querySelector('.newsletter-form');
    if (!form) return;

    form.addEventListener('submit', async function(e) {
        e.preventDefault();

        const email = this.querySelector('input[type="email"]').value;

        try {
            const response = await fetch('/api/newsletter', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ email })
            });

            if (response.ok) {
                showNotification('Successfully subscribed to newsletter!');
                this.reset();
            } else {
                throw new Error('Failed to subscribe');
            }
        } catch (error) {
            showNotification('Error subscribing to newsletter', 'error');
        }
    });
}

// Featured Posts Slider
function initializeFeaturedSlider() {
    const slider = document.querySelector('.featured-slider');
    if (!slider) return;

    new Swiper('.featured-slider', {
        slidesPerView: 1,
        spaceBetween: 30,
        loop: true,
        autoplay: {
            delay: 5000,
            disableOnInteraction: false,
        },
        pagination: {
            el: '.swiper-pagination',
            clickable: true,
        },
        navigation: {
            nextEl: '.swiper-button-next',
            prevEl: '.swiper-button-prev',
        },
        breakpoints: {
            768: {
                slidesPerView: 2,
            },
            1024: {
                slidesPerView: 3,
            },
        }
    });
}

// Utility Functions
function showNotification(message, type = 'success') {
    const notification = document.createElement('div');
    notification.className = `notification ${type}`;
    notification.textContent = message;

    document.body.appendChild(notification);

    setTimeout(() => {
        notification.classList.add('show');
        setTimeout(() => {
            notification.classList.remove('show');
            setTimeout(() => notification.remove(), 300);
        }, 3000);
    }, 100);
}

function formatDate(dateString) {
    const options = { year: 'numeric', month: 'long', day: 'numeric' };
    return new Date(dateString).toLocaleDateString(undefined, options);
}

// Initialize animations
function initializeAnimations() {
    const observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                entry.target.classList.add('animate');
                observer.unobserve(entry.target);
            }
        });
    }, {
        threshold: 0.1
    });

    document.querySelectorAll('.animate-on-scroll').forEach(el => {
        observer.observe(el);
    });
}