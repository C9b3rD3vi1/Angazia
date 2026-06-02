var AngaziaWS = (function () {
  var WS_URL = (window.location.protocol === 'https:' ? 'wss://' : 'ws://')
    + window.location.host + '/ws';

  var socket = null;
  var token = null;
  var connected = false;
  var reconnectAttempts = 0;
  var maxReconnectAttempts = 10;
  var reconnectDelay = 1000;
  var maxReconnectDelay = 30000;
  var reconnectTimer = null;
  var heartbeatTimer = null;
  var heartbeatInterval = 30000;
  var heartbeatTimeout = null;
  var heartbeatTimeoutDuration = 10000;
  var listeners = {};
  var messageQueue = [];
  var intentionalClose = false;
  var clientId = null;

  function on(evt, fn) {
    if (!listeners[evt]) listeners[evt] = [];
    listeners[evt].push(fn);
    return function () {
      listeners[evt] = (listeners[evt] || []).filter(function (f) { return f !== fn; });
    };
  }

  function off(evt, fn) {
    if (!listeners[evt]) return;
    if (fn) listeners[evt] = listeners[evt].filter(function (f) { return f !== fn; });
    else delete listeners[evt];
  }

  function emit(evt, payload) {
    (listeners[evt] || []).forEach(function (fn) {
      try { fn(payload); } catch (e) { console.error('[WS] listener error:', evt, e); }
    });
  }

  function connect(authToken) {
    token = authToken;
    intentionalClose = false;

    if (socket && (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING)) {
      return;
    }

    try {
      var url = WS_URL;
      if (token) url += '?token=' + encodeURIComponent(token);
      socket = new WebSocket(url);
    } catch (e) {
      console.error('[WS] connection failed:', e);
      scheduleReconnect();
      return;
    }

    socket.onopen = function () {
      connected = true;
      reconnectAttempts = 0;
      reconnectDelay = 1000;
      emit('connected', { clientId: clientId });
      startHeartbeat();
      flushQueue();
    };

    socket.onmessage = function (event) {
      try {
        var msg = JSON.parse(event.data);
        handleMessage(msg);
      } catch (e) {
        console.error('[WS] parse error:', e);
      }
    };

    socket.onclose = function (event) {
      connected = false;
      stopHeartbeat();
      emit('disconnected', { code: event.code, reason: event.reason });
      if (!intentionalClose && reconnectAttempts < maxReconnectAttempts) {
        scheduleReconnect();
      }
    };

    socket.onerror = function (err) {
      console.error('[WS] error:', err);
      emit('error', err);
    };
  }

  function disconnect() {
    intentionalClose = true;
    clearTimeout(reconnectTimer);
    reconnectAttempts = maxReconnectAttempts;
    stopHeartbeat();
    if (socket) {
      socket.onclose = null;
      socket.close(1000, 'Client disconnect');
      socket = null;
    }
    connected = false;
    emit('disconnected', { code: 1000, reason: 'client disconnect' });
  }

  function send(type, payload) {
    var msg = JSON.stringify({ type: type, payload: payload || {} });
    if (socket && socket.readyState === WebSocket.OPEN) {
      socket.send(msg);
      return true;
    }
    messageQueue.push(msg);
    return false;
  }

  function flushQueue() {
    while (messageQueue.length > 0) {
      var msg = messageQueue.shift();
      if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(msg);
      }
    }
  }

  function handleMessage(msg) {
    switch (msg.type) {
      case 'connected':
        clientId = msg.payload && msg.payload.client_id;
        emit('connected', msg.payload);
        break;

      case 'notification':
        emit('notification', msg.payload);
        if (window.AngaziaNotifications) {
          AngaziaNotifications.onNotificationReceived(msg.payload);
        }
        break;

      case 'ping':
        send('pong');
        break;

      case 'pong':
        clearTimeout(heartbeatTimeout);
        break;

      case 'message':
        emit('message', msg.payload);
        break;

      case 'typing':
        emit('typing', msg.payload);
        break;

      case 'error':
        console.error('[WS] server error:', msg.payload);
        emit('error', msg.payload);
        break;

      default:
        emit(msg.type, msg.payload);
        break;
    }
  }

  function startHeartbeat() {
    stopHeartbeat();
    heartbeatTimer = setInterval(function () {
      send('ping');
      heartbeatTimeout = setTimeout(function () {
        console.warn('[WS] heartbeat timeout, reconnecting...');
        if (socket) {
          socket.onclose = null;
          socket.close();
          socket = null;
        }
        connected = false;
        scheduleReconnect();
      }, heartbeatTimeoutDuration);
    }, heartbeatInterval);
  }

  function stopHeartbeat() {
    if (heartbeatTimer) { clearInterval(heartbeatTimer); heartbeatTimer = null; }
    if (heartbeatTimeout) { clearTimeout(heartbeatTimeout); heartbeatTimeout = null; }
  }

  function scheduleReconnect() {
    if (intentionalClose || reconnectAttempts >= maxReconnectAttempts) return;
    var delay = Math.min(reconnectDelay * Math.pow(2, reconnectAttempts), maxReconnectDelay);
    var jitter = Math.random() * 1000;
    reconnectAttempts++;
    clearTimeout(reconnectTimer);
    reconnectTimer = setTimeout(function () {
      if (!intentionalClose) connect(token);
    }, delay + jitter);
  }

  function isConnected() { return connected; }
  function getClientId() { return clientId; }

  return {
    connect: connect,
    disconnect: disconnect,
    send: send,
    on: on,
    off: off,
    isConnected: isConnected,
    getClientId: getClientId,
    reconnectNow: function () {
      reconnectAttempts = 0;
      if (!connected) connect(token);
    },
  };
})();
