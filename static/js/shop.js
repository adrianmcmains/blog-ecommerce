// Initialize shop functionality when DOM is loaded
document.addEventListener('DOMContentLoaded', function() {
    initShop();
});

// Cart state
let cart = {
    items: [],
    total: 0
};

function initShop() {
    // Load cart from localStorage
    loadCart();
    
    // Initialize event listeners
    initEventListeners();
    
    // Initialize components
    initializeQuickView();
    initProductGallery();
    initAnimations();
}

function initEventListeners() {
    // Add to cart buttons
    document.querySelectorAll('.add-to-cart').forEach(button => {
        button.addEventListener('click', function() {
            const id = this.dataset.id;
            const title = this.dataset.title;
            const price = parseFloat(this.dataset.price);
            addToCart(id, title, price);
        });
    });

    // Quick view buttons
    document.querySelectorAll('.quick-view').forEach(button => {
        button.addEventListener('click', function() {
            const productId = this.dataset.productId;
            const title = this.dataset.productTitle;
            const price = this.dataset.productPrice;
            const description = this.dataset.productDescription;
            showQuickView(productId, title, price, description);
        });
    });
}

// Cart Functions
function loadCart() {
    const savedCart = localStorage.getItem('cart');
    if (savedCart) {
        cart = JSON.parse(savedCart);
        updateCartDisplay();
    }
}

function saveCart() {
    localStorage.setItem('cart', JSON.stringify(cart));
    updateCartDisplay();
}

function addToCart(id, title, price) {
    // Find if product already exists in cart
    const existingItem = cart.items.find(item => item.id === id);
    
    if (existingItem) {
        existingItem.quantity += 1;
    } else {
        cart.items.push({
            id: id,
            title: title,
            price: price,
            quantity: 1
        });
    }
    
    // Update cart total
    updateCartTotal();
    
    // Save cart
    saveCart();
    
    // Show notification
    showNotification(`Added ${title} to cart`);
    
    // Animate cart icon
    animateCartIcon();
}

function removeFromCart(id) {
    cart.items = cart.items.filter(item => item.id !== id);
    updateCartTotal();
    saveCart();
}

function updateQuantity(id, quantity) {
    const item = cart.items.find(item => item.id === id);
    if (item) {
        item.quantity = parseInt(quantity);
        if (item.quantity <= 0) {
            removeFromCart(id);
        } else {
            updateCartTotal();
            saveCart();
        }
    }
}

function updateCartTotal() {
    cart.total = cart.items.reduce((sum, item) => sum + (item.price * item.quantity), 0);
}

function updateCartDisplay() {
    // Update cart count
    const cartCount = document.querySelector('.cart-count');
    if (cartCount) {
        const totalItems = cart.items.reduce((sum, item) => sum + item.quantity, 0);
        cartCount.textContent = totalItems;
    }

    // Update cart total
    const cartTotal = document.querySelector('.cart-total');
    if (cartTotal) {
        cartTotal.textContent = `$${cart.total.toFixed(2)}`;
    }

    // Update cart items list if it exists
    const cartList = document.querySelector('.cart-items-list');
    if (cartList) {
        cartList.innerHTML = cart.items.map(item => `
            <div class="cart-item" data-id="${item.id}">
                <div class="cart-item-title">${item.title}</div>
                <div class="cart-item-price">$${(item.price * item.quantity).toFixed(2)}</div>
                <div class="cart-item-quantity">
                    <button class="btn btn-sm btn-outline-secondary" onclick="updateQuantity('${item.id}', ${item.quantity - 1})">-</button>
                    <span>${item.quantity}</span>
                    <button class="btn btn-sm btn-outline-secondary" onclick="updateQuantity('${item.id}', ${item.quantity + 1})">+</button>
                </div>
                <button class="btn btn-sm btn-danger" onclick="removeFromCart('${item.id}')">Remove</button>
            </div>
        `).join('');
    }
}

// Quick View Functions
function initializeQuickView() {
    const quickViewModal = document.getElementById('quickViewModal');
    if (quickViewModal) {
        quickViewModal.addEventListener('hidden.bs.modal', function() {
            const content = quickViewModal.querySelector('.quick-view-content');
            if (content) {
                content.innerHTML = '';
            }
        });
    }
}

function showQuickView(productId, title, price, description) {
    const modal = document.getElementById('quickViewModal');
    const content = modal.querySelector('.quick-view-content');
    
    content.innerHTML = `
        <div class="row">
            <div class="col-md-6">
                <img src="/images/products/${productId}.jpg" class="img-fluid" alt="${title}">
            </div>
            <div class="col-md-6">
                <h3>${title}</h3>
                <p class="price">$${price}</p>
                <p class="description">${description}</p>
                <div class="d-grid gap-2">
                    <button class="btn btn-primary" onclick="addToCart('${productId}', '${title}', ${price})">
                        Add to Cart
                    </button>
                </div>
            </div>
        </div>
    `;
}

// Product Gallery Functions
function initProductGallery() {
    const mainImage = document.querySelector('.gallery-main-image');
    const thumbnails = document.querySelectorAll('.gallery-thumbnail');

    if (mainImage && thumbnails.length > 0) {
        thumbnails.forEach(thumb => {
            thumb.addEventListener('click', function() {
                // Update main image
                mainImage.src = this.dataset.image;
                // Update active state
                thumbnails.forEach(t => t.classList.remove('active'));
                this.classList.add('active');
            });
        });
    }
}

// Animation Functions
function initAnimations() {
    // Initialize animation observer
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

    // Observe all animated elements
    document.querySelectorAll('.animate-on-scroll').forEach(el => {
        observer.observe(el);
    });
}

function animateCartIcon() {
    const cartIcon = document.querySelector('.cart-icon');
    if (cartIcon) {
        cartIcon.classList.add('bounce');
        setTimeout(() => cartIcon.classList.remove('bounce'), 300);
    }
}

// Utility Functions
function showNotification(message) {
    // Remove existing notifications
    const existingNotifications = document.querySelectorAll('.shop-notification');
    existingNotifications.forEach(notification => notification.remove());

    // Create new notification
    const notification = document.createElement('div');
    notification.className = 'shop-notification';
    notification.innerHTML = `
        <div class="notification-content">
            <i class="bi bi-check-circle-fill"></i>
            <span>${message}</span>
        </div>
    `;

    // Add to document
    document.body.appendChild(notification);

    // Trigger animation
    setTimeout(() => notification.classList.add('show'), 10);

    // Remove after delay
    setTimeout(() => {
        notification.classList.remove('show');
        setTimeout(() => notification.remove(), 300);
    }, 3000);
}

// Export functions that need to be accessed globally
window.addToCart = addToCart;
window.removeFromCart = removeFromCart;
window.updateQuantity = updateQuantity;