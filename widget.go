package nimschatwidget

import "net/http"

func handleWidget(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Write([]byte(widgetJS))
}

const widgetJS = `(function() {
  var cfg = window.nimschatwidgetConfig || {};
  var basePath = cfg.baseURL || cfg.basePath || '/admin/chat';
  var sessionId = cfg.sessionId || 'default';
  var defaultNim = cfg.defaultNim || '';
  function getContext() { var c = window.nimschatwidgetConfig || {}; return c.context || ''; }

  // Prevent double-init
  if (document.getElementById('ncw-root')) return;

  // --- Styles (aligned with nimsforestwebchat design) ---
  var style = document.createElement('style');
  style.textContent = [
    '@keyframes ncw-msg-appear { from { opacity:0;transform:translateY(8px) scale(0.97); } to { opacity:1;transform:translateY(0) scale(1); } }',
    '@keyframes ncw-typing-bounce { 0%,60%,100% { transform:translateY(0); } 30% { transform:translateY(-6px); } }',
    '#ncw-root { font-size:15px; }',
    '#ncw-btn { position:fixed;bottom:24px;right:24px;width:56px;height:56px;border-radius:50%;background:#4AA847;border:none;cursor:pointer;box-shadow:0 4px 12px rgba(0,0,0,.25);z-index:99999;display:flex;align-items:center;justify-content:center;transition:transform .15s cubic-bezier(0.34,1.56,0.64,1),background .15s; }',
    '#ncw-btn:hover { background:#3d8f3c;transform:scale(1.08); }',
    '#ncw-btn:active { transform:scale(0.95); }',
    '#ncw-btn svg { width:28px;height:28px;fill:#fff; }',
    '#ncw-panel { position:fixed;top:0;right:-420px;width:400px;height:100%;background:#F8FAF5;box-shadow:-4px 0 24px rgba(0,0,0,.15);z-index:100000;display:flex;flex-direction:column;transition:right .25s cubic-bezier(0.16,1,0.3,1); }',
    '#ncw-panel.ncw-open { right:0; }',
    '#ncw-header { display:flex;align-items:center;gap:12px;padding:0 12px;height:56px;background:#fff;color:#1E3A1C;border-bottom:1px solid #EDE9E5;flex-shrink:0; }',
    '#ncw-header select { flex:1;padding:6px 8px;border-radius:8px;border:1px solid #E2DDD8;background:#F0F3ED;color:#1E3A1C;font-family:Inter,DM Sans,-apple-system,BlinkMacSystemFont,sans-serif;font-size:13px;font-weight:600;appearance:auto;outline:none; }',
    '#ncw-header select:focus { border-color:#A8D5A2; }',
    '#ncw-close { width:40px;height:40px;display:flex;align-items:center;justify-content:center;background:none;border:none;border-radius:9999px;color:#1E3A1C;font-size:22px;cursor:pointer;transition:background .15s; }',
    '#ncw-close:hover { background:#F0F3ED; }',
    '#ncw-involve-human { width:40px;height:40px;display:flex;align-items:center;justify-content:center;background:none;border:none;border-radius:9999px;color:#1E3A1C;font-size:18px;cursor:pointer;transition:background .15s;margin-left:auto; }',
    '#ncw-involve-human:hover { background:#F0F3ED; }',
    '#ncw-involve-human[disabled] { opacity:0.4;cursor:not-allowed; }',
    '#ncw-messages { flex:1;overflow-y:auto;padding:16px;display:flex;flex-direction:column;gap:8px;background:#F8FAF5; }',
    '.ncw-msg-wrap { display:flex;flex-direction:column;animation:ncw-msg-appear .25s cubic-bezier(0.16,1,0.3,1) both; }',
    '.ncw-msg-wrap-user { align-items:flex-end; }',
    '.ncw-msg-wrap-nim { align-items:flex-start; }',
    '.ncw-msg { max-width:85%;padding:8px 12px;font-family:Georgia,"Source Serif 4",serif;font-size:15px;line-height:1.5;word-wrap:break-word;white-space:pre-wrap; }',
    '.ncw-msg-user { background:#DCF5DB;color:#1E3A1C;border-radius:16px 16px 4px 16px; }',
    '.ncw-msg-nim { background:#fff;color:#6B5B4E;border:1px solid #EDE9E5;border-radius:16px 16px 16px 4px; }',
    '.ncw-msg-error { background:#fee2e2;color:#991b1b;border:1px solid #fca5a5;border-radius:16px 16px 16px 4px; }',
    '.ncw-nim-name { font-family:Inter,DM Sans,-apple-system,BlinkMacSystemFont,sans-serif;font-size:11px;color:#8A7D73;font-weight:600;margin-bottom:2px; }',
    '.ncw-typing { align-self:flex-start;display:flex;align-items:center;gap:4px;padding:8px 12px;background:#fff;border:1px solid #EDE9E5;border-radius:16px 16px 16px 4px; }',
    '.ncw-typing-dot { width:8px;height:8px;border-radius:50%;background:#A8D5A2;animation:ncw-typing-bounce 1.4s ease-in-out infinite; }',
    '.ncw-typing-dot:nth-child(2) { animation-delay:0.2s; }',
    '.ncw-typing-dot:nth-child(3) { animation-delay:0.4s; }',
    '#ncw-input-area { display:flex;align-items:flex-end;gap:8px;padding:8px 12px;border-top:1px solid #EDE9E5;background:#fff;flex-shrink:0; }',
    '#ncw-input { flex:1;padding:8px 12px;border:1px solid #E2DDD8;border-radius:12px;background:#F0F3ED;font-family:Georgia,"Source Serif 4",serif;font-size:15px;color:#6B5B4E;resize:none;line-height:1.4;min-height:40px;max-height:120px;outline:none;transition:border-color .2s; }',
    '#ncw-input:focus { border-color:#A8D5A2; }',
    '#ncw-input::placeholder { color:#B0A89F; }',
    '#ncw-send { width:40px;height:40px;border-radius:50%;background:#4AA847;border:none;cursor:pointer;display:flex;align-items:center;justify-content:center;flex-shrink:0;transition:transform .15s cubic-bezier(0.34,1.56,0.64,1); }',
    '#ncw-send:hover { transform:scale(1.05); }',
    '#ncw-send:active { transform:scale(0.9); }',
    '#ncw-send:disabled { opacity:0.4;cursor:not-allowed;transform:none; }',
    '#ncw-send svg { width:20px;height:20px;fill:#fff; }',
    'body.ncw-panel-open { overflow:hidden;position:fixed;width:100%;height:100%; }',
    '@media (max-width:768px) { #ncw-panel { width:100%;right:-100%;top:0;bottom:0;height:100vh;height:100dvh; } #ncw-panel.ncw-open { right:0; } #ncw-btn { bottom:16px;right:16px;width:48px;height:48px; } #ncw-btn svg { width:24px;height:24px; } #ncw-header { padding:0 8px;padding-top:env(safe-area-inset-top,0);height:auto;min-height:48px; } #ncw-header select { font-size:14px; } #ncw-messages { padding:12px 8px; } #ncw-input { font-size:16px; } #ncw-input-area { padding:8px;padding-bottom:max(8px,env(safe-area-inset-bottom,8px)); } }'
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

  // --- Panel ---
  var panel = document.createElement('div');
  panel.id = 'ncw-panel';
  panel.innerHTML = [
    '<div id="ncw-header">',
    '  <select id="ncw-nim-select"><option value="">Loading nims...</option></select>',
    '  <button id="ncw-involve-human" title="Involve a human">&#128100;</button>',
    '  <button id="ncw-close">&times;</button>',
    '</div>',
    '<div id="ncw-messages"></div>',
    '<div id="ncw-input-area">',
    '  <textarea id="ncw-input" rows="1" placeholder="Type a message..."></textarea>',
    '  <button id="ncw-send"><svg viewBox="0 0 24 24"><path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z"/></svg></button>',
    '</div>'
  ].join('');
  root.appendChild(panel);

  document.body.appendChild(root);

  // --- Refs ---
  var nimSelect = document.getElementById('ncw-nim-select');
  var messages = document.getElementById('ncw-messages');
  var input = document.getElementById('ncw-input');
  var sendBtn = document.getElementById('ncw-send');
  var closeBtn = document.getElementById('ncw-close');
  var involveHumanBtn = document.getElementById('ncw-involve-human');
  var isOpen = false;
  var eventSource = null;
  var sending = false;

  // --- Toggle ---
  var scrollY = 0;
  function openPanel() {
    isOpen = true;
    scrollY = window.scrollY;
    document.body.style.top = -scrollY + 'px';
    document.body.classList.add('ncw-panel-open');
    panel.classList.add('ncw-open');
    connectSSE();
    input.focus();
  }
  function closePanel() {
    isOpen = false;
    panel.classList.remove('ncw-open');
    document.body.classList.remove('ncw-panel-open');
    document.body.style.top = '';
    window.scrollTo(0, scrollY);
  }
  btn.addEventListener('click', function() {
    if (isOpen) closePanel();
    else openPanel();
  });
  closeBtn.addEventListener('click', closePanel);
  involveHumanBtn.addEventListener('click', function() {
    if (involveHumanBtn.disabled) return;
    var nim = nimSelect.value || 'nimble';
    involveHumanBtn.disabled = true;
    addMessage('Requesting a human to join...', 'nim', nim);
    fetch(basePath + '/send', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'same-origin',
      body: JSON.stringify({
        session_id: sessionId,
        target_nim: nim,
        text: '[involve_human]',
        context: getContext()
      })
    }).then(function(r) {
      if (!r.ok) throw new Error('server returned ' + r.status);
    }).catch(function(e) {
      addMessage('Failed to involve human: ' + e.message, 'error', '');
      involveHumanBtn.disabled = false;
    });
  });

  // --- Load nims ---
  fetch(basePath + '/nims')
    .then(function(r) { return r.json(); })
    .then(function(nims) {
      nimSelect.innerHTML = '';
      nims.forEach(function(n, i) {
        var opt = document.createElement('option');
        opt.value = n.name;
        opt.textContent = n.name + ' — ' + n.role;
        if (defaultNim && n.name === defaultNim) opt.selected = true;
        else if (!defaultNim && i === 0) opt.selected = true;
        nimSelect.appendChild(opt);
      });
    })
    .catch(function() {
      nimSelect.innerHTML = '<option value="nimble">nimble — General assistant</option>';
    });

  // --- Send ---
  function sendMessage() {
    var text = input.value.trim();
    if (!text || sending) return;

    var nim = nimSelect.value || 'nimble';
    addMessage(text, 'user', '');
    input.value = '';
    autoResize();
    sending = true;
    sendBtn.disabled = true;

    showTyping();

    fetch(basePath + '/send', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'same-origin',
      body: JSON.stringify({
        session_id: sessionId,
        target_nim: nim,
        text: text,
        context: getContext()
      })
    })
    .then(function(r) {
      if (!r.ok) throw new Error('send failed');
    })
    .catch(function(e) {
      hideTyping();
      addMessage('Failed to send message: ' + e.message, 'error', '');
    })
    .finally(function() {
      sending = false;
      sendBtn.disabled = false;
    });
  }

  sendBtn.addEventListener('click', sendMessage);
  input.addEventListener('keydown', function(e) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    }
  });

  // --- Auto-resize textarea ---
  function autoResize() {
    input.style.height = 'auto';
    input.style.height = Math.min(input.scrollHeight, 120) + 'px';
  }
  input.addEventListener('input', autoResize);

  // --- Messages ---
  function addMessage(text, type, nimName) {
    hideTyping();
    var wrap = document.createElement('div');
    wrap.className = 'ncw-msg-wrap ncw-msg-wrap-' + type;
    var bubble = document.createElement('div');
    bubble.className = 'ncw-msg ncw-msg-' + type;
    if (type === 'nim' && nimName) {
      var nameEl = document.createElement('div');
      nameEl.className = 'ncw-nim-name';
      nameEl.textContent = nimName;
      bubble.appendChild(nameEl);
    }
    var textEl = document.createElement('span');
    textEl.textContent = text;
    bubble.appendChild(textEl);
    wrap.appendChild(bubble);
    messages.appendChild(wrap);
    messages.scrollTop = messages.scrollHeight;
  }

  var typingEl = null;
  function showTyping() {
    if (typingEl) return;
    typingEl = document.createElement('div');
    typingEl.className = 'ncw-typing';
    typingEl.innerHTML = '<div class="ncw-typing-dot"></div><div class="ncw-typing-dot"></div><div class="ncw-typing-dot"></div>';
    messages.appendChild(typingEl);
    messages.scrollTop = messages.scrollHeight;
  }
  function hideTyping() {
    if (typingEl) {
      typingEl.remove();
      typingEl = null;
    }
  }

  // --- SSE ---
  function connectSSE() {
    if (eventSource) return;
    eventSource = new EventSource(basePath + '/events?session=' + encodeURIComponent(sessionId));
    eventSource.onmessage = function(e) {
      try {
        var msg = JSON.parse(e.data);
        hideTyping();
        if (msg.is_error) {
          addMessage(msg.text, 'error', msg.source);
        } else {
          addMessage(msg.text, 'nim', msg.source);
        }
      } catch(err) {}
    };
    eventSource.onerror = function() {
      eventSource.close();
      eventSource = null;
      // Auto-reconnect after 3s if panel is open
      if (isOpen) {
        setTimeout(function() {
          if (isOpen) connectSSE();
        }, 3000);
      }
    };
  }
})();
`
