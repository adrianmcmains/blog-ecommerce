/**
 * Shopping Cart Management
 * 
 * This script handles the client-side state management for the shopping cart.
 * It uses local storage to persist cart data between page refreshes.
 */

class ShoppingCart {
  constructor() {
    this.items = [];
    this.totalItems = 0;
    this.totalPrice = 0;
    this.init();
  }

  /**
   * Initialize the cart from local storage
   */
  init() {
    const savedCart = localStorage.getItem('shoppingCart');
    if (savedCart) {
      try {
        const cartData = JSON.parse(savedCart);
        this.items = cartData.items || [];
        this.recalculate();
      } catch (error) {
        console.error('Error parsing cart data from localStorage', error);
        this.clearCart();
      }
    }
    this.updateCartDisplay();
  }

  /**
   * Add an item to the cart
   * @param {Object} product - Product to add
   * @param {Number} quantity - Quantity to add
   */
  addItem(product, quantity = 1) {
    const existingItemIndex = this.items.findIndex(item => item.id === product.id);
    
    if (existingItemIndex >= 0) {
      // Update existing item quantity
      this.items[existingItemIndex].quantity += quantity;
    } else {
      // Add new item
      this.items.push({
        id: product.id,
        name: product.name,
        price: product.price,
        image: product.image,
        quantity: quantity
      });
    }
    
    this.saveCart();
    this.recalculate();
    this.updateCartDisplay();
    
    // Show notification
    this.showNotification(`Added ${product.name} to cart`);
  }

  /**
   * Update an item's quantity
   * @param {Number} productId - Product ID to update
   * @param {Number} quantity - New quantity
   */
  updateQuantity(productId, quantity) {
    const itemIndex = this.items.findIndex(item => item.id === productId);
    
    if (itemIndex >= 0) {
      if (quantity <= 0) {
        // Remove item if quantity is zero or negative
        this.removeItem(productId);
      } else {
        // Update quantity
        this.items[itemIndex].quantity = quantity;
        this.saveCart();
        this.recalculate();
        this.updateCartDisplay();
      }
    }
  }

  /**
   * Remove an item from the cart
   * @param {Number} productId - Product ID to remove
   */
  removeItem(productId) {
    this.items = this.items.filter(item => item.id !== productId);
    this.saveCart();
    this.recalculate();
    this.updateCartDisplay();
  }

  /**
   * Clear the entire cart
   */
  clearCart() {
    this.items = [];
    this.saveCart();
    this.recalculate();
    this.updateCartDisplay();
  }

  /**
   * Recalculate cart totals
   */
  recalculate() {
    this.totalItems = this.items.reduce((total, item) => total + item.quantity, 0);
    this.totalPrice = this.items.reduce((total, item) => total + (item.price * item.quantity), 0);
  }

  /**
   * Save cart to local storage
   */
  saveCart() {
    localStorage.setItem('shoppingCart', JSON.stringify({
      items: this.items,
      lastUpdated: new Date().toISOString()
    }));
  }

  /**
   * Show notification when adding to cart
   * @param {String} message - Notification message
   */
  showNotification(message) {
    const notification = document.createElement('div');
    notification.className = 'cart-notification';
    notification.innerHTML = `<p>${message}</p>`;
    document.body.appendChild(notification);
    
    // Auto-remove after 3 seconds
    setTimeout(() => {
      notification.classList.add('fadeOut');
      setTimeout(() => {
        document.body.removeChild(notification);
      }, 500);
    }, 3000);
  }

  /**
   * Update cart UI elements
   */
  updateCartDisplay() {
    // Update cart counter
    const cartCounters = document.querySelectorAll('.cart-counter');
    cartCounters.forEach(counter => {
      counter.textContent = this.totalItems;
    });
    
    // Update cart items in dropdown or cart page if they exist
    const cartItemsContainer = document.getElementById('cart-items');
    if (cartItemsContainer) {
      this.renderCartItems(cartItemsContainer);
    }
    
    // Update cart total if it exists
    const cartTotal = document.getElementById('cart-total');
    if (cartTotal) {
      cartTotal.textContent = `$${this.totalPrice.toFixed(2)}`;
    }
  }

  /**
   * Render cart items into a container
   * @param {HTMLElement} container - Container to render items into
   */
  renderCartItems(container) {
    // Clear the container
    container.innerHTML = '';
    
    if (this.items.length === 0) {
      container.innerHTML = '<p class="empty-cart-message">Your cart is empty</p>';
      return;
    }
    
    // Create items list
    const itemsList = document.createElement('ul');
    itemsList.className = 'cart-items-list';
    
    this.items.forEach(item => {
      const itemElement = document.createElement('li');
      itemElement.className = 'cart-item';
      itemElement.innerHTML = `
        <div class="cart-item-image">
          <img src="${item.image}" alt="${item.name}">
        </div>
        <div class="cart-item-details">
          <h4 class="cart-item-name">${item.name}</h4>
          <div class="cart-item-price">${(item.price).toFixed(2)}</div>
          <div class="cart-item-quantity">
            <button class="quantity-btn decrease" data-id="${item.id}">-</button>
            <input type="number" value="${item.quantity}" min="1" max="99" data-id="${item.id}" class="quantity-input">
            <button class="quantity-btn increase" data-id="${item.id}">+</button>
          </div>
        </div>
        <div class="cart-item-total">
          ${(item.price * item.quantity).toFixed(2)}
        </div>
        <button class="remove-item-btn" data-id="${item.id}">
          <i class="fas fa-trash"></i>
        </button>
      `;
      
      itemsList.appendChild(itemElement);
    });
    
    container.appendChild(itemsList);
    
    // Add subtotal and checkout button
    const cartSummary = document.createElement('div');
    cartSummary.className = 'cart-summary';
    cartSummary.innerHTML = `
      <div class="cart-subtotal">
        <span>Subtotal:</span>
        <span class="subtotal-amount">${this.totalPrice.toFixed(2)}</span>
      </div>
      <button id="checkout-btn" class="btn btn-primary">Proceed to Checkout</button>
    `;
    
    container.appendChild(cartSummary);
    
    // Add event listeners
    this.addCartEventListeners();
  }
  
  /**
   * Add event listeners to cart elements
   */
  addCartEventListeners() {
    // Quantity decrease buttons
    document.querySelectorAll('.quantity-btn.decrease').forEach(btn => {
      btn.addEventListener('click', e => {
        const productId = parseInt(e.target.dataset.id);
        const item = this.items.find(item => item.id === productId);
        if (item && item.quantity > 1) {
          this.updateQuantity(productId, item.quantity - 1);
        }
      });
    });
    
    // Quantity increase buttons
    document.querySelectorAll('.quantity-btn.increase').forEach(btn => {
      btn.addEventListener('click', e => {
        const productId = parseInt(e.target.dataset.id);
        const item = this.items.find(item => item.id === productId);
        if (item) {
          this.updateQuantity(productId, item.quantity + 1);
        }
      });
    });
    
    // Quantity input fields
    document.querySelectorAll('.quantity-input').forEach(input => {
      input.addEventListener('change', e => {
        const productId = parseInt(e.target.dataset.id);
        const newQuantity = parseInt(e.target.value);
        if (!isNaN(newQuantity) && newQuantity > 0) {
          this.updateQuantity(productId, newQuantity);
        }
      });
    });
    
    // Remove item buttons
    document.querySelectorAll('.remove-item-btn').forEach(btn => {
      btn.addEventListener('click', e => {
        const productId = parseInt(e.target.closest('.remove-item-btn').dataset.id);
        this.removeItem(productId);
      });
    });
    
    // Checkout button
    const checkoutBtn = document.getElementById('checkout-btn');
    if (checkoutBtn) {
      checkoutBtn.addEventListener('click', () => {
        window.location.href = '/checkout';
      });
    }
  }
  
  /**
   * Sync cart with the server (for logged-in users)
   * @param {String} userId - User ID to sync cart with
   */
  async syncWithServer(userId) {
    try {
      // First get server cart data
      const response = await fetch(`/api/cart/${userId}`);
      if (response.ok) {
        const serverCart = await response.json();
        
        // Merge server cart with local cart
        this.mergeWithServerCart(serverCart);
        
        // Update server with merged cart
        await this.updateServerCart(userId);
      }
    } catch (error) {
      console.error('Error syncing cart with server', error);
    }
  }
  
  /**
   * Merge server cart with local cart
   * @param {Object} serverCart - Cart data from server
   */
  mergeWithServerCart(serverCart) {
    if (!serverCart || !serverCart.items || !Array.isArray(serverCart.items)) {
      return;
    }
    
    // For each server item, add it to the local cart if it doesn't exist
    serverCart.items.forEach(serverItem => {
      const localItem = this.items.find(item => item.id === serverItem.id);
      
      if (!localItem) {
        // Add server item to local cart
        this.items.push(serverItem);
      } else {
        // Update local item with server item data if server item is newer
        if (new Date(serverItem.updatedAt) > new Date(localItem.updatedAt)) {
          localItem.quantity = serverItem.quantity;
        }
      }
    });
    
    this.saveCart();
    this.recalculate();
    this.updateCartDisplay();
  }
  
  /**
   * Update server with current cart data
   * @param {String} userId - User ID to update cart for
   */
  async updateServerCart(userId) {
    try {
      const response = await fetch(`/api/cart/${userId}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('authToken')}`
        },
        body: JSON.stringify({
          items: this.items
        })
      });
      
      if (!response.ok) {
        console.error('Error updating server cart');
      }
    } catch (error) {
      console.error('Error updating server cart', error);
    }
  }
}

// Initialize cart when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
  // Create global cart instance
  window.cart = new ShoppingCart();
  
  // Add to cart buttons
  document.querySelectorAll('.add-to-cart-btn').forEach(btn => {
    btn.addEventListener('click', e => {
      e.preventDefault();
      
      const productId = parseInt(e.target.dataset.id);
      const productName = e.target.dataset.name;
      const productPrice = parseFloat(e.target.dataset.price);
      const productImage = e.target.dataset.image;
      const quantity = parseInt(e.target.dataset.quantity || 1);
      
      // Add item to cart
      window.cart.addItem({
        id: productId,
        name: productName,
        price: productPrice,
        image: productImage
      }, quantity);
    });
  });
  
  // Check if user is logged in and sync cart
  const authToken = localStorage.getItem('authToken');
  if (authToken) {
    fetch('/api/auth/me', {
      headers: {
        'Authorization': `Bearer ${authToken}`
      }
    })
    .then(response => {
      if (response.ok) {
        return response.json();
      }
      throw new Error('Not authenticated');
    })
    .then(userData => {
      window.cart.syncWithServer(userData.id);
    })
    .catch(error => {
      console.log('User not authenticated, using local cart only');
    });
  }
});