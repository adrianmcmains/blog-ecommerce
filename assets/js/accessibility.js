/**
 * Accessibility Utility
 * 
 * This script enhances the website's accessibility by providing:
 * - Keyboard navigation support
 * - Focus management
 * - Screen reader improvements
 * - High contrast mode
 * - Font size adjustments
 */

class AccessibilityHelper {
    constructor() {
      this.isHighContrastMode = false;
      this.fontSizeLevel = 2; // Normal (levels: 0-small, 1-medium-small, 2-normal, 3-medium-large, 4-large)
      this.init();
    }
  
    /**
     * Initialize accessibility features
     */
    init() {
      this.initKeyboardNavigation();
      this.initSkipLinks();
      this.initARIA();
      this.initAccessibilityMenu();
      this.initFocusManagement();
      
      // Check for saved preferences
      this.loadUserPreferences();
    }
  
    /**
     * Initialize keyboard navigation
     */
    initKeyboardNavigation() {
      // Add tabindex to interactive elements that may not have it
      document.querySelectorAll('.interactive-element').forEach(el => {
        if (!el.hasAttribute('tabindex')) {
          el.setAttribute('tabindex', '0');
        }
      });
  
      // Enhance hover menus to be keyboard accessible
      document.querySelectorAll('.dropdown').forEach(dropdown => {
        const trigger = dropdown.querySelector('.dropdown-toggle');
        const menu = dropdown.querySelector('.dropdown-menu');
        
        if (trigger && menu) {
          trigger.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' || e.key === ' ') {
              e.preventDefault();
              const expanded = trigger.getAttribute('aria-expanded') === 'true';
              trigger.setAttribute('aria-expanded', !expanded);
              menu.classList.toggle('show');
              
              if (!expanded) {
                // Focus the first menu item when opening
                const firstItem = menu.querySelector('a');
                if (firstItem) firstItem.focus();
              }
            }
          });
          
          // Allow ESC key to close dropdown
          menu.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') {
              trigger.setAttribute('aria-expanded', 'false');
              menu.classList.remove('show');
              trigger.focus();
            }
          });
        }
      });
    }
  
    /**
     * Initialize skip links for keyboard users
     */
    initSkipLinks() {
      const skipLink = document.createElement('a');
      skipLink.href = '#main';
      skipLink.className = 'skip-link';
      skipLink.textContent = 'Skip to main content';
      
      document.body.insertBefore(skipLink, document.body.firstChild);
      
      // Make sure the main content is focusable
      const mainContent = document.getElementById('main');
      if (mainContent && !mainContent.hasAttribute('tabindex')) {
        mainContent.setAttribute('tabindex', '-1');
      }
    }
  
    /**
     * Initialize proper ARIA attributes
     */
    initARIA() {
      // Set page regions
      const header = document.querySelector('header');
      if (header) header.setAttribute('role', 'banner');
      
      const main = document.getElementById('main');
      if (main) main.setAttribute('role', 'main');
      
      const footer = document.querySelector('footer');
      if (footer) footer.setAttribute('role', 'contentinfo');
      
      const nav = document.querySelector('nav');
      if (nav) nav.setAttribute('role', 'navigation');
      
      // Set search role
      const searchForm = document.querySelector('form[role="search"]');
      if (searchForm) {
        const searchInput = searchForm.querySelector('input');
        if (searchInput) searchInput.setAttribute('aria-label', 'Search');
      }
      
      // Set button controls
      document.querySelectorAll('button').forEach(button => {
        if (!button.textContent.trim() && !button.getAttribute('aria-label')) {
          // Button has no text, check for an icon and use that for the label
          const icon = button.querySelector('i.fa, i.fas, i.far, i.fab');
          if (icon) {
            const iconClass = Array.from(icon.classList)
              .find(cls => cls.startsWith('fa-'));
            
            if (iconClass) {
              const iconName = iconClass.replace('fa-', '');
              button.setAttribute('aria-label', iconName);
            }
          }
        }
      });
      
      // Add labels to form fields without labels
      document.querySelectorAll('input, select, textarea').forEach(field => {
        if (!field.id) return;
        
        // Check if field has an associated label
        const hasLabel = document.querySelector(`label[for="${field.id}"]`);
        
        if (!hasLabel && !field.getAttribute('aria-label')) {
          // Try to find a placeholder or name to use as label
          const labelText = field.placeholder || field.name;
          if (labelText) {
            field.setAttribute('aria-label', labelText);
          }
        }
      });
    }
  
    /**
     * Initialize accessibility menu
     */
    initAccessibilityMenu() {
      // Create accessibility menu
      const accessibilityMenu = document.createElement('div');
      accessibilityMenu.className = 'accessibility-menu';
      accessibilityMenu.innerHTML = `
        <button aria-label="Accessibility Options" class="accessibility-toggle">
          <i class="fas fa-universal-access"></i>
        </button>
        <div class="accessibility-panel">
          <h3>Accessibility Options</h3>
          <div class="accessibility-option">
            <button id="high-contrast-toggle">
              <i class="fas fa-adjust"></i> High Contrast
            </button>
          </div>
          <div class="accessibility-option">
            <button id="font-size-decrease">
              <i class="fas fa-font fa-xs"></i> Smaller Text
            </button>
            <button id="font-size-reset">
              <i class="fas fa-font"></i> Reset Text
            </button>
            <button id="font-size-increase">
              <i class="fas fa-font fa-lg"></i> Larger Text
            </button>
          </div>
        </div>
      `;
      
      document.body.appendChild(accessibilityMenu);
      
      // Add event listeners
      const toggleButton = accessibilityMenu.querySelector('.accessibility-toggle');
      const panel = accessibilityMenu.querySelector('.accessibility-panel');
      
      toggleButton.addEventListener('click', () => {
        panel.classList.toggle('active');
        const expanded = panel.classList.contains('active');
        toggleButton.setAttribute('aria-expanded', expanded);
      });
      
      // High contrast toggle
      const highContrastToggle = document.getElementById('high-contrast-toggle');
      highContrastToggle.addEventListener('click', () => {
        this.toggleHighContrast();
      });
      
      // Font size controls
      document.getElementById('font-size-decrease').addEventListener('click', () => {
        this.changeFontSize(-1);
      });
      
      document.getElementById('font-size-reset').addEventListener('click', () => {
        this.resetFontSize();
      });
      
      document.getElementById('font-size-increase').addEventListener('click', () => {
        this.changeFontSize(1);
      });
    }
  
    /**
     * Initialize focus management
     */
    initFocusManagement() {
      // Add focus indicator styles
      const style = document.createElement('style');
      style.textContent = `
        :focus {
          outline: 2px solid #1e90ff !important;
          outline-offset: 2px !important;
        }
        
        .skip-link {
          position: absolute;
          top: -40px;
          left: 0;
          padding: 8px 16px;
          background-color: #ffffff;
          color: #0066cc;
          z-index: 1001;
          transition: top 0.2s;
        }
        
        .skip-link:focus {
          top: 0;
        }
        
        .accessibility-menu {
          position: fixed;
          bottom: 20px;
          right: 20px;
          z-index: 1000;
        }
        
        .accessibility-toggle {
          width: 48px;
          height: 48px;
          border-radius: 50%;
          background-color: #0066cc;
          color: white;
          border: none;
          cursor: pointer;
          box-shadow: 0 2px 5px rgba(0,0,0,0.2);
        }
        
        .accessibility-panel {
          position: absolute;
          bottom: 60px;
          right: 0;
          width: 250px;
          background-color: white;
          border-radius: 5px;
          padding: 15px;
          box-shadow: 0 2px 10px rgba(0,0,0,0.2);
          display: none;
        }
        
        .accessibility-panel.active {
          display: block;
        }
        
        .accessibility-option {
          margin: 10px 0;
        }
        
        .accessibility-option button {
          background: none;
          border: 1px solid #ccc;
          padding: 5px 10px;
          border-radius: 3px;
          cursor: pointer;
          margin: 2px;
        }
        
        .accessibility-option button:hover {
          background-color: #f0f0f0;
        }
        
        /* High contrast mode */
        body.high-contrast {
          background-color: #000 !important;
          color: #fff !important;
        }
        
        body.high-contrast a {
          color: #ffff00 !important;
        }
        
        body.high-contrast button, 
        body.high-contrast .btn {
          background-color: #000 !important;
          color: #fff !important;
          border: 2px solid #fff !important;
        }
        
        body.high-contrast input, 
        body.high-contrast select, 
        body.high-contrast textarea {
          background-color: #000 !important;
          color: #fff !important;
          border: 1px solid #fff !important;
        }
        
        /* Font size adjustments */
        body.font-size-0 {
          font-size: 80% !important;
        }
        
        body.font-size-1 {
          font-size: 90% !important;
        }
        
        body.font-size-2 {
          font-size: 100% !important;
        }
        
        body.font-size-3 {
          font-size: 120% !important;
        }
        
        body.font-size-4 {
          font-size: 150% !important;
        }
      `;
      
      document.head.appendChild(style);
      
      // Trap focus in modals
      document.querySelectorAll('.modal').forEach(modal => {
        modal.addEventListener('keydown', (e) => {
          if (e.key !== 'Tab') return;
          
          const focusableElements = modal.querySelectorAll(
            'a[href], button, textarea, input, select, [tabindex]:not([tabindex="-1"])'
          );
          
          if (focusableElements.length === 0) return;
          
          const firstElement = focusableElements[0];
          const lastElement = focusableElements[focusableElements.length - 1];
          
          if (e.shiftKey && document.activeElement === firstElement) {
            e.preventDefault();
            lastElement.focus();
          } else if (!e.shiftKey && document.activeElement === lastElement) {
            e.preventDefault();
            firstElement.focus();
          }
        });
      });
    }
  
    /**
     * Toggle high contrast mode
     */
    toggleHighContrast() {
      document.body.classList.toggle('high-contrast');
      this.isHighContrastMode = document.body.classList.contains('high-contrast');
      this.saveUserPreferences();
    }
  
    /**
     * Change font size
     * @param {Number} direction - 1 to increase, -1 to decrease
     */
    changeFontSize(direction) {
      // Remove current font size class
      document.body.classList.remove(`font-size-${this.fontSizeLevel}`);
      
      // Update font size level
      this.fontSizeLevel = Math.max(0, Math.min(4, this.fontSizeLevel + direction));
      
      // Add new font size class
      document.body.classList.add(`font-size-${this.fontSizeLevel}`);
      
      // Save preferences
      this.saveUserPreferences();
    }
  
    /**
     * Reset font size to normal
     */
    resetFontSize() {
      // Remove current font size class
      document.body.classList.remove(`font-size-${this.fontSizeLevel}`);
      
      // Reset to normal
      this.fontSizeLevel = 2;
      
      // Add new font size class
      document.body.classList.add(`font-size-${this.fontSizeLevel}`);
      
      // Save preferences
      this.saveUserPreferences();
    }
  
    /**
     * Save user preferences to localStorage
     */
    saveUserPreferences() {
      try {
        localStorage.setItem('accessibility', JSON.stringify({
          highContrast: this.isHighContrastMode,
          fontSizeLevel: this.fontSizeLevel
        }));
      } catch (e) {
        console.error('Failed to save accessibility preferences:', e);
      }
    }
  
    /**
     * Load user preferences from localStorage
     */
    loadUserPreferences() {
      try {
        const preferences = JSON.parse(localStorage.getItem('accessibility'));
        
        if (preferences) {
          // Apply high contrast if enabled
          if (preferences.highContrast) {
            document.body.classList.add('high-contrast');
            this.isHighContrastMode = true;
          }
          
          // Apply font size
          if (typeof preferences.fontSizeLevel === 'number') {
            document.body.classList.remove(`font-size-${this.fontSizeLevel}`);
            this.fontSizeLevel = Math.max(0, Math.min(4, preferences.fontSizeLevel));
            document.body.classList.add(`font-size-${this.fontSizeLevel}`);
          }
        }
      } catch (e) {
        console.error('Failed to load accessibility preferences:', e);
      }
    }
  }
  
  // Initialize accessibility helper when DOM is loaded
  document.addEventListener('DOMContentLoaded', () => {
    window.accessibilityHelper = new AccessibilityHelper();
  });