/**
 * Migro Documentation - Main JavaScript
 */

(function() {
  'use strict';

  // Initialize all modules when DOM is ready
  document.addEventListener('DOMContentLoaded', function() {
    initSidebar();
    initCodeCopy();
    initSmoothScroll();
    initBackToTop();
  });

  /**
   * Sidebar Module
   */
  function initSidebar() {
    const menuToggle = document.querySelector('.menu-toggle');
    const sidebar = document.querySelector('.sidebar');
    const overlay = document.querySelector('.overlay');
    const navItems = document.querySelectorAll('.nav-item.has-submenu');

    // Mobile menu toggle
    if (menuToggle) {
      menuToggle.addEventListener('click', function() {
        document.body.classList.toggle('sidebar-open');
      });
    }

    // Close sidebar when clicking overlay
    if (overlay) {
      overlay.addEventListener('click', function() {
        document.body.classList.remove('sidebar-open');
      });
    }

    // Submenu toggle
    navItems.forEach(function(item) {
      const link = item.querySelector('.nav-link');
      if (link) {
        link.addEventListener('click', function(e) {
          // Only toggle if clicking on the arrow or if it's a submenu-only item
          const hasHref = link.getAttribute('href') && link.getAttribute('href') !== '#';
          const clickedArrow = e.target.classList.contains('nav-arrow');

          if (clickedArrow || !hasHref) {
            e.preventDefault();
            item.classList.toggle('expanded');
          }
        });
      }
    });

    // Highlight current page
    highlightCurrentPage();

    // Close sidebar on escape key
    document.addEventListener('keydown', function(e) {
      if (e.key === 'Escape' && document.body.classList.contains('sidebar-open')) {
        document.body.classList.remove('sidebar-open');
      }
    });
  }

  /**
   * Highlight current page in navigation
   */
  function highlightCurrentPage() {
    const currentPath = window.location.pathname.split('/').pop() || 'index.html';
    const currentHash = window.location.hash;
    const navLinks = document.querySelectorAll('.nav-link');

    navLinks.forEach(function(link) {
      const href = link.getAttribute('href');
      if (!href) return;

      const linkPath = href.split('#')[0];
      const linkHash = href.includes('#') ? '#' + href.split('#')[1] : '';

      // Check if this link matches current page
      if (linkPath === currentPath || (currentPath === '' && linkPath === 'index.html')) {
        // For submenu items, also check hash
        if (linkHash && currentHash) {
          if (linkHash === currentHash) {
            link.classList.add('active');
          }
        } else if (!linkHash) {
          link.classList.add('active');
          // Expand parent submenu if exists
          const parentItem = link.closest('.nav-item.has-submenu');
          if (parentItem) {
            parentItem.classList.add('expanded');
          }
        }
      }
    });

    // Expand submenu containing active link
    const activeSubmenuLink = document.querySelector('.nav-submenu .nav-link.active');
    if (activeSubmenuLink) {
      const parentItem = activeSubmenuLink.closest('.nav-item.has-submenu');
      if (parentItem) {
        parentItem.classList.add('expanded');
      }
    }
  }

  /**
   * Code Copy Module
   */
  function initCodeCopy() {
    // Find all code blocks and add copy buttons
    const codeBlocks = document.querySelectorAll('.code-block');

    codeBlocks.forEach(function(block) {
      const copyBtn = block.querySelector('.code-copy');
      const codeElement = block.querySelector('code');

      if (copyBtn && codeElement) {
        copyBtn.addEventListener('click', function() {
          copyToClipboard(codeElement.textContent, copyBtn);
        });
      }
    });

    // Also handle standalone pre elements
    const standalonePres = document.querySelectorAll('pre:not(.code-block pre)');
    standalonePres.forEach(function(pre) {
      // Create copy button
      const copyBtn = document.createElement('button');
      copyBtn.className = 'code-copy standalone';
      copyBtn.textContent = 'Copy';
      copyBtn.style.cssText = 'position: absolute; top: 8px; right: 8px;';

      // Make pre relative for positioning
      pre.style.position = 'relative';
      pre.appendChild(copyBtn);

      const codeElement = pre.querySelector('code') || pre;
      copyBtn.addEventListener('click', function() {
        copyToClipboard(codeElement.textContent, copyBtn);
      });
    });
  }

  /**
   * Copy text to clipboard
   */
  function copyToClipboard(text, button) {
    // Try modern clipboard API first
    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(text).then(function() {
        showCopiedFeedback(button);
      }).catch(function() {
        fallbackCopy(text, button);
      });
    } else {
      fallbackCopy(text, button);
    }
  }

  /**
   * Fallback copy method for older browsers
   */
  function fallbackCopy(text, button) {
    const textarea = document.createElement('textarea');
    textarea.value = text;
    textarea.style.cssText = 'position: fixed; left: -9999px;';
    document.body.appendChild(textarea);
    textarea.select();

    try {
      document.execCommand('copy');
      showCopiedFeedback(button);
    } catch (err) {
      console.error('Failed to copy:', err);
    }

    document.body.removeChild(textarea);
  }

  /**
   * Show copied feedback on button
   */
  function showCopiedFeedback(button) {
    const originalText = button.textContent;
    button.textContent = 'Copied!';
    button.classList.add('copied');

    setTimeout(function() {
      button.textContent = originalText;
      button.classList.remove('copied');
    }, 2000);
  }

  /**
   * Smooth Scroll Module
   */
  function initSmoothScroll() {
    // Handle anchor links
    document.querySelectorAll('a[href^="#"]').forEach(function(anchor) {
      anchor.addEventListener('click', function(e) {
        const targetId = this.getAttribute('href');
        if (targetId === '#') return;

        const targetElement = document.querySelector(targetId);
        if (targetElement) {
          e.preventDefault();

          // Close mobile sidebar if open
          document.body.classList.remove('sidebar-open');

          // Scroll to target
          targetElement.scrollIntoView({
            behavior: 'smooth',
            block: 'start'
          });

          // Update URL hash
          history.pushState(null, null, targetId);
        }
      });
    });
  }

  /**
   * Back to Top Button Module
   */
  function initBackToTop() {
    const backToTopBtn = document.querySelector('.back-to-top');
    if (!backToTopBtn) return;

    // Show/hide button based on scroll position
    function toggleBackToTop() {
      if (window.scrollY > 300) {
        backToTopBtn.classList.add('visible');
      } else {
        backToTopBtn.classList.remove('visible');
      }
    }

    // Throttle scroll event
    let ticking = false;
    window.addEventListener('scroll', function() {
      if (!ticking) {
        window.requestAnimationFrame(function() {
          toggleBackToTop();
          ticking = false;
        });
        ticking = true;
      }
    });

    // Scroll to top on click
    backToTopBtn.addEventListener('click', function() {
      window.scrollTo({
        top: 0,
        behavior: 'smooth'
      });
    });

    // Initial check
    toggleBackToTop();
  }

})();
