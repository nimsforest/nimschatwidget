package nimschatwidget

const widgetJS = `(function() {
  // Prevent double-init
  if (document.getElementById('ncw-root')) return;

  function getWebchatURL() { return (window.nimschatwidgetConfig || {}).webchatURL || ''; }
  function getContext() { return (window.nimschatwidgetConfig || {}).context || ''; }

  // Session ID: persist in localStorage
  var sessionId = localStorage.getItem('ncw-session');
  if (!sessionId) {
    sessionId = 'ncw-' + Math.random().toString(36).substring(2) + Date.now().toString(36);
    localStorage.setItem('ncw-session', sessionId);
  }

  // Expand state: persist in localStorage
  var isExpanded = localStorage.getItem('ncw-expanded') === 'true';

  // --- Styles ---
  var style = document.createElement('style');
  style.textContent = [
    '#ncw-root { font-size:15px; }',
    '#ncw-btn { position:fixed;bottom:24px;right:24px;width:56px;height:56px;border-radius:50%;background:#4AA847;border:none;cursor:pointer;box-shadow:0 4px 12px rgba(0,0,0,.25);z-index:99999;display:flex;align-items:center;justify-content:center;transition:transform .15s cubic-bezier(0.34,1.56,0.64,1),background .15s; }',
    '#ncw-btn:hover { background:#3d8f3c;transform:scale(1.08); }',
    '#ncw-btn:active { transform:scale(0.95); }',
    '#ncw-btn svg { width:28px;height:28px;fill:#fff; }',
    '#ncw-iframe-panel { position:fixed;bottom:96px;right:24px;width:400px;height:600px;border:none;border-radius:16px;box-shadow:0 8px 32px rgba(0,0,0,0.2);z-index:99998;display:none;overflow:hidden;background:#F8FAF5;transition:all .25s cubic-bezier(0.16,1,0.3,1); }',
    '#ncw-iframe-panel.ncw-open { display:block; }',
    '#ncw-iframe-panel.ncw-expanded { bottom:16px;right:16px;width:calc(100vw - 32px);height:calc(100vh - 32px);max-width:800px;max-height:900px;border-radius:16px; }',
    '#ncw-iframe { width:100%;height:100%;border:none;border-radius:16px; }',
    '#ncw-close { position:absolute;top:8px;right:8px;width:32px;height:32px;display:flex;align-items:center;justify-content:center;background:rgba(255,255,255,0.9);border:none;border-radius:50%;color:#1E3A1C;font-size:18px;cursor:pointer;z-index:99999;box-shadow:0 2px 8px rgba(0,0,0,0.15);transition:background .15s; }',
    '#ncw-close:hover { background:#fff; }',
    '#ncw-expand { position:absolute;top:8px;right:48px;width:32px;height:32px;display:flex;align-items:center;justify-content:center;background:rgba(255,255,255,0.9);border:none;border-radius:50%;color:#1E3A1C;cursor:pointer;z-index:99999;box-shadow:0 2px 8px rgba(0,0,0,0.15);transition:background .15s; }',
    '#ncw-expand:hover { background:#fff; }',
    '#ncw-expand svg { width:16px;height:16px; }',
    '@media (max-width:640px) { #ncw-iframe-panel { top:0;left:0;right:0;bottom:0;width:100%;height:100%;border-radius:0; } #ncw-iframe-panel.ncw-expanded { top:0;left:0;right:0;bottom:0;width:100%;height:100%;max-width:100%;max-height:100%;border-radius:0; } #ncw-iframe { border-radius:0; } #ncw-btn { bottom:16px;right:16px;width:48px;height:48px; } #ncw-btn svg { width:24px;height:24px; } }'
  ].join('\n');
  document.head.appendChild(style);

  // --- Root ---
  var root = document.createElement('div');
  root.id = 'ncw-root';

  // --- FAB ---
  var btn = document.createElement('button');
  btn.id = 'ncw-btn';
  btn.title = 'Chat with a nim';
  btn.innerHTML = '<svg viewBox="0 0 24 24"><path d="M20 2H4c-1.1 0-2 .9-2 2v18l4-4h14c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zm0 14H5.2L4 17.2V4h16v12z"/></svg>';
  root.appendChild(btn);

  // --- Iframe panel ---
  var panel = document.createElement('div');
  panel.id = 'ncw-iframe-panel';
  panel.style.position = 'fixed';

  var closeBtn = document.createElement('button');
  closeBtn.id = 'ncw-close';
  closeBtn.innerHTML = '&times;';
  panel.appendChild(closeBtn);

  var expandBtn = document.createElement('button');
  expandBtn.id = 'ncw-expand';
  panel.appendChild(expandBtn);

  var iframe = document.createElement('iframe');
  iframe.id = 'ncw-iframe';
  iframe.allow = 'clipboard-write';
  panel.appendChild(iframe);

  root.appendChild(panel);
  document.body.appendChild(root);

  // --- Expand/collapse icons ---
  var expandIcon = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="15 3 21 3 21 9"/><polyline points="9 21 3 21 3 15"/><line x1="21" y1="3" x2="14" y2="10"/><line x1="3" y1="21" x2="10" y2="14"/></svg>';
  var collapseIcon = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="4 14 10 14 10 20"/><polyline points="20 10 14 10 14 4"/><line x1="14" y1="10" x2="21" y2="3"/><line x1="3" y1="21" x2="10" y2="14"/></svg>';

  function updateExpandButton() {
    expandBtn.innerHTML = isExpanded ? collapseIcon : expandIcon;
    expandBtn.title = isExpanded ? 'Compact view' : 'Expand view';
    if (isExpanded) {
      panel.classList.add('ncw-expanded');
    } else {
      panel.classList.remove('ncw-expanded');
    }
  }
  updateExpandButton();

  // --- State ---
  var isOpen = false;

  function buildIframeSrc() {
    var url = getWebchatURL();
    if (!url) return '';
    return url + '/embed?session=' + encodeURIComponent(sessionId) + '&context=' + encodeURIComponent(getContext());
  }

  function openPanel() {
    var src = buildIframeSrc();
    if (!src) {
      console.error('[nimschatwidget] webchatURL not configured');
      return;
    }
    if (!iframe.src || iframe.src === 'about:blank') {
      iframe.src = src;
    } else {
      // Send updated context to existing iframe
      try {
        iframe.contentWindow.postMessage({type: 'context', context: getContext()}, '*');
      } catch(e) {}
    }
    isOpen = true;
    panel.classList.add('ncw-open');
  }

  function closePanel() {
    isOpen = false;
    panel.classList.remove('ncw-open');
  }

  btn.addEventListener('click', function() {
    if (isOpen) closePanel();
    else openPanel();
  });

  closeBtn.addEventListener('click', function(e) {
    e.stopPropagation();
    closePanel();
  });

  expandBtn.addEventListener('click', function(e) {
    e.stopPropagation();
    isExpanded = !isExpanded;
    localStorage.setItem('ncw-expanded', isExpanded ? 'true' : 'false');
    updateExpandButton();
  });
})();
`
