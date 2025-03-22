/**
 * Eversend Payment Integration
 * 
 * This script handles the client-side integration with the Eversend payment gateway.
 */

class PaymentService {
    constructor(apiUrl) {
      this.apiUrl = apiUrl || '/api';
      this.authToken = localStorage.getItem('authToken');
    }
  
    /**
     * Set the authentication token
     * @param {String} token - JWT authentication token
     */
    setAuthToken(token) {
      this.authToken = token;
      localStorage.setItem('authToken', token);
    }
  
    /**
     * Initiate a payment for an order
     * @param {Number} orderId - Order ID to pay for
     * @param {String} currency - Currency code (e.g., 'USD')
     * @param {String} redirectUrl - URL to redirect after payment
     * @param {Object} metadata - Additional metadata to include
     * @returns {Promise} Promise resolving to payment information
     */
    async initiatePayment(orderId, currency, redirectUrl, metadata = {}) {
      try {
        if (!this.authToken) {
          throw new Error('Authentication required to process payment');
        }
  
        const response = await fetch(`${this.apiUrl}/payments/initiate`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${this.authToken}`
          },
          body: JSON.stringify({
            orderId,
            currency,
            redirectUrl,
            metadata
          })
        });
  
        if (!response.ok) {
          // Try to get error message from response
          let errorMessage = 'Failed to initiate payment';
          try {
            const errorData = await response.json();
            errorMessage = errorData.error || errorMessage;
          } catch (e) {
            // Ignore parsing error
          }
          throw new Error(errorMessage);
        }
  
        return response.json();
      } catch (error) {
        console.error('Payment initiation error:', error);
        throw error;
      }
    }
  
    /**
     * Get the status of a payment
     * @param {String} paymentId - Payment ID to check
     * @returns {Promise} Promise resolving to payment status
     */
    async getPaymentStatus(paymentId) {
      try {
        if (!this.authToken) {
          throw new Error('Authentication required to check payment');
        }
  
        const response = await fetch(`${this.apiUrl}/payments/${paymentId}`, {
          method: 'GET',
          headers: {
            'Authorization': `Bearer ${this.authToken}`
          }
        });
  
        if (!response.ok) {
          // Try to get error message from response
          let errorMessage = 'Failed to get payment status';
          try {
            const errorData = await response.json();
            errorMessage = errorData.error || errorMessage;
          } catch (e) {
            // Ignore parsing error
          }
          throw new Error(errorMessage);
        }
  
        return response.json();
      } catch (error) {
        console.error('Payment status error:', error);
        throw error;
      }
    }
  
    /**
     * Cancel a payment
     * @param {String} paymentId - Payment ID to cancel
     * @returns {Promise} Promise resolving to cancellation result
     */
    async cancelPayment(paymentId) {
      try {
        if (!this.authToken) {
          throw new Error('Authentication required to cancel payment');
        }
  
        const response = await fetch(`${this.apiUrl}/payments/${paymentId}/cancel`, {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${this.authToken}`
          }
        });
  
        if (!response.ok) {
          // Try to get error message from response
          let errorMessage = 'Failed to cancel payment';
          try {
            const errorData = await response.json();
            errorMessage = errorData.error || errorMessage;
          } catch (e) {
            // Ignore parsing error
          }
          throw new Error(errorMessage);
        }
  
        return response.json();
      } catch (error) {
        console.error('Payment cancellation error:', error);
        throw error;
      }
    }
  
    /**
     * Process a successful payment return
     * @param {String} paymentId - Payment ID that was processed
     * @returns {Promise} Promise resolving to payment result
     */
    async processPaymentReturn(paymentId) {
      try {
        // Get the payment status
        const paymentStatus = await this.getPaymentStatus(paymentId);
        
        // Process based on status
        if (paymentStatus.status === 'completed') {
          return {
            success: true,
            message: 'Payment completed successfully',
            paymentStatus
          };
        } else if (paymentStatus.status === 'pending') {
          return {
            success: false,
            message: 'Payment is still processing. We will notify you when it completes.',
            paymentStatus
          };
        } else {
          return {
            success: false,
            message: `Payment failed with status: ${paymentStatus.status}`,
            paymentStatus
          };
        }
      } catch (error) {
        console.error('Payment return processing error:', error);
        throw error;
      }
    }
  }
  
  // Checkout page functionality
  document.addEventListener('DOMContentLoaded', () => {
    // Initialize payment service
    const paymentService = new PaymentService();
    
    // Get the checkout form
    const checkoutForm = document.getElementById('checkout-form');
    if (!checkoutForm) return;
    
    // Handle checkout form submission
    checkoutForm.addEventListener('submit', async (e) => {
      e.preventDefault();
      
      try {
        // Show loading state
        const submitButton = checkoutForm.querySelector('button[type="submit"]');
        const originalButtonText = submitButton.textContent;
        submitButton.disabled = true;
        submitButton.textContent = 'Processing...';
        
        // Get form data
        const formData = new FormData(checkoutForm);
        
        // First, create the order via API
        const orderResponse = await fetch('/api/shop/orders', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${localStorage.getItem('authToken')}`
          },
          body: JSON.stringify({
            shippingAddress: formData.get('shipping_address'),
            billingAddress: formData.get('billing_address'),
            paymentMethod: 'eversend',
            notes: formData.get('notes') || ''
          })
        });
        
        if (!orderResponse.ok) {
          throw new Error('Failed to create order');
        }
        
        const orderData = await orderResponse.json();
        
        // Now initiate payment for this order
        const currentUrl = window.location.origin;
        const redirectUrl = `${currentUrl}/checkout/confirmation`;
        
        const paymentResponse = await paymentService.initiatePayment(
          orderData.id,
          'USD', // Currency could be configurable
          redirectUrl
        );
        
        if (paymentResponse.success) {
          // Redirect to Eversend payment page
          window.location.href = paymentResponse.paymentUrl;
        } else {
          throw new Error(paymentResponse.errorMessage || 'Payment initiation failed');
        }
      } catch (error) {
        console.error('Checkout error:', error);
        
        // Show error message
        const errorElement = document.getElementById('checkout-error');
        if (errorElement) {
          errorElement.textContent = error.message;
          errorElement.style.display = 'block';
        } else {
          alert(`Error: ${error.message}`);
        }
        
        // Reset button
        const submitButton = checkoutForm.querySelector('button[type="submit"]');
        submitButton.disabled = false;
        submitButton.textContent = originalButtonText;
      }
    });
    
    // Handle payment return (on confirmation page)
    const confirmationPage = document.getElementById('payment-confirmation');
    if (confirmationPage) {
      const urlParams = new URLSearchParams(window.location.search);
      const paymentId = urlParams.get('payment_id');
      const status = urlParams.get('status');
      
      if (paymentId) {
        (async () => {
          try {
            // Show loading message
            const messageElement = document.getElementById('confirmation-message');
            if (messageElement) {
              messageElement.textContent = 'Verifying payment status...';
            }
            
            // Process the payment return
            const result = await paymentService.processPaymentReturn(paymentId);
            
            // Update UI based on result
            if (messageElement) {
              messageElement.textContent = result.message;
              messageElement.className = result.success ? 'success-message' : 'error-message';
            }
            
            // If payment was successful, clear the cart
            if (result.success && window.cart) {
              window.cart.clearCart();
            }
            
            // Show order details if available
            const orderDetailsElement = document.getElementById('order-details');
            if (orderDetailsElement && result.paymentStatus) {
              // This would typically fetch order details from the server
              // and populate them in the page
              const orderDetailUrl = `/api/shop/orders/${result.paymentStatus.metadata.orderID}`;
              
              try {
                const orderResponse = await fetch(orderDetailUrl, {
                  headers: {
                    'Authorization': `Bearer ${localStorage.getItem('authToken')}`
                  }
                });
                
                if (orderResponse.ok) {
                  const orderData = await orderResponse.json();
                  // Render order details (implementation depends on your UI)
                  renderOrderDetails(orderData, orderDetailsElement);
                }
              } catch (orderError) {
                console.error('Error fetching order details:', orderError);
              }
            }
          } catch (error) {
            console.error('Payment confirmation error:', error);
            
            // Show error message
            const messageElement = document.getElementById('confirmation-message');
            if (messageElement) {
              messageElement.textContent = `Error: ${error.message}`;
              messageElement.className = 'error-message';
            }
          }
        })();
      } else if (status === 'canceled') {
        // Handle canceled payment
        const messageElement = document.getElementById('confirmation-message');
        if (messageElement) {
          messageElement.textContent = 'Payment was canceled.';
          messageElement.className = 'info-message';
        }
      }
    }
  });
  
  // Helper function to render order details
  function renderOrderDetails(order, container) {
    // Clear container
    container.innerHTML = '';
    
    // Create order summary
    const orderSummary = document.createElement('div');
    orderSummary.className = 'order-summary';
    orderSummary.innerHTML = `
      <h3>Order Summary</h3>
      <p><strong>Order Number:</strong> #${order.id}</p>
      <p><strong>Date:</strong> ${new Date(order.createdAt).toLocaleDateString()}</p>
      <p><strong>Status:</strong> <span class="order-status ${order.status}">${order.status}</span></p>
      <p><strong>Total:</strong> $${order.totalAmount.toFixed(2)}</p>
    `;
    
    // Create items table
    const itemsTable = document.createElement('table');
    itemsTable.className = 'order-items-table';
    
    // Add table header
    itemsTable.innerHTML = `
      <thead>
        <tr>
          <th>Product</th>
          <th>Quantity</th>
          <th>Price</th>
          <th>Total</th>
        </tr>
      </thead>
      <tbody></tbody>
    `;
    
    // Add table body
    const tableBody = itemsTable.querySelector('tbody');
    order.items.forEach(item => {
      const row = document.createElement('tr');
      row.innerHTML = `
        <td>${item.productName}</td>
        <td>${item.quantity}</td>
        <td>$${item.unitPrice.toFixed(2)}</td>
        <td>$${item.totalPrice.toFixed(2)}</td>
      `;
      tableBody.appendChild(row);
    });
    
    // Add shipping info
    const shippingInfo = document.createElement('div');
    shippingInfo.className = 'shipping-info';
    shippingInfo.innerHTML = `
      <h3>Shipping Information</h3>
      <p>${order.shippingAddress.replace(/\n/g, '<br>')}</p>
    `;
    
    // Append all elements to container
    container.appendChild(orderSummary);
    container.appendChild(itemsTable);
    container.appendChild(shippingInfo);
  }