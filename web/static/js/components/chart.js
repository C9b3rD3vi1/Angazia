var AngaziaChart = (function () {
  var accent = '#00e5a0';
  var textColor = '#111827';
  var mutedColor = '#6b7280';
  var gridColor = '#e5e7eb';
  var dangerColor = '#ef4444';
  var warningColor = '#f59e0b';

  function getColors(opts) {
    var c = {};
    if (opts && opts.colors && opts.colors.length) c.palette = opts.colors;
    else c.palette = ['#00e5a0', '#3b82f6', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899', '#14b8a6', '#f97316'];
    c.text = (opts && opts.textColor) || textColor;
    c.muted = (opts && opts.mutedColor) || mutedColor;
    c.grid = (opts && opts.gridColor) || gridColor;
    return c;
  }

  function chartId() {
    return 'ac-' + Math.random().toString(36).substring(2, 9);
  }

  function sanitizeVal(v) {
    var n = Number(v);
    return isNaN(n) ? 0 : n;
  }

  function line(container, data, opts) {
    if (typeof container === 'string') container = document.querySelector(container);
    if (!container || !data || !data.length) { renderEmpty(container); return; }
    opts = opts || {};
    var colors = getColors(opts);

    var svgNS = 'http://www.w3.org/2000/svg';
    var id = chartId();
    var w = opts.width || container.clientWidth || 400;
    var h = opts.height || 240;
    var pad = opts.padding || { top: 20, right: 20, bottom: 30, left: 40 };
    if (!opts.padding) { pad.top = 20; pad.right = 20; pad.bottom = 30; pad.left = 40; }

    var seriesKeys = Object.keys(data[0]).filter(function (k) { return k !== (opts.xKey || 'label') && typeof data[0][k] === 'number'; });
    var xKey = opts.xKey || 'label';
    var seriesColors = opts.seriesColors || {};
    var curved = opts.curved !== false;

    var plotW = w - pad.left - pad.right;
    var plotH = h - pad.top - pad.bottom;

    var allVals = [];
    data.forEach(function (d) {
      seriesKeys.forEach(function (k) { allVals.push(sanitizeVal(d[k])); });
    });
    var minVal = opts.minY !== undefined ? opts.minY : Math.min(0, Math.min.apply(null, allVals));
    var maxVal = opts.maxY !== undefined ? opts.maxY : Math.max.apply(null, allVals);
    var range = maxVal - minVal || 1;

    var xStep = data.length > 1 ? plotW / (data.length - 1) : plotW;

    function xPos(i) { return pad.left + (data.length > 1 ? i * xStep : plotW / 2); }
    function yPos(v) { return pad.top + plotH - ((sanitizeVal(v) - minVal) / range) * plotH; }

    var svg = createSvg(w, h, id);

    if (opts.showGrid !== false) {
      var gridLines = 5;
      for (var gi = 0; gi <= gridLines; gi++) {
        var gy = pad.top + (plotH / gridLines) * gi;
        var lineEl = document.createElementNS(svgNS, 'line');
        lineEl.setAttribute('x1', pad.left);
        lineEl.setAttribute('y1', gy);
        lineEl.setAttribute('x2', w - pad.right);
        lineEl.setAttribute('y2', gy);
        lineEl.setAttribute('stroke', colors.grid);
        lineEl.setAttribute('stroke-width', '1');
        lineEl.setAttribute('stroke-dasharray', '4,4');
        svg.appendChild(lineEl);
        var gv = maxVal - (range / gridLines) * gi;
        var txt = document.createElementNS(svgNS, 'text');
        txt.setAttribute('x', pad.left - 8);
        txt.setAttribute('y', gy + 4);
        txt.setAttribute('text-anchor', 'end');
        txt.setAttribute('fill', colors.muted);
        txt.setAttribute('font-size', '10');
        txt.textContent = formatNum(gv);
        svg.appendChild(txt);
      }
    }

    seriesKeys.forEach(function (key, si) {
      var color = seriesColors[key] || colors.palette[si % colors.palette.length];
      var points = data.map(function (d, i) {
        return xPos(i) + ',' + yPos(d[key]);
      });

      if (curved) {
        var path = buildSmoothPath(points, 0.3);
        var pEl = document.createElementNS(svgNS, 'path');
        pEl.setAttribute('d', path);
        pEl.setAttribute('fill', 'none');
        pEl.setAttribute('stroke', color);
        pEl.setAttribute('stroke-width', '2.5');
        pEl.setAttribute('stroke-linecap', 'round');
        pEl.setAttribute('stroke-linejoin', 'round');
        svg.appendChild(pEl);
      } else {
        var poly = document.createElementNS(svgNS, 'polyline');
        poly.setAttribute('points', points.join(' '));
        poly.setAttribute('fill', 'none');
        poly.setAttribute('stroke', color);
        poly.setAttribute('stroke-width', '2.5');
        poly.setAttribute('stroke-linecap', 'round');
        poly.setAttribute('stroke-linejoin', 'round');
        svg.appendChild(poly);
      }

      data.forEach(function (d, i) {
        var cx = xPos(i);
        var cy = yPos(d[key]);
        var dot = document.createElementNS(svgNS, 'circle');
        dot.setAttribute('cx', cx);
        dot.setAttribute('cy', cy);
        dot.setAttribute('r', '3.5');
        dot.setAttribute('fill', '#fff');
        dot.setAttribute('stroke', color);
        dot.setAttribute('stroke-width', '2.5');
        dot.style.cursor = 'pointer';
        dot.addEventListener('mouseenter', function (e) {
          showTooltip(e, key + ': ' + formatNum(d[key]) + (data[i][xKey] ? ' (' + data[i][xKey] + ')' : ''));
        });
        dot.addEventListener('mouseleave', hideTooltip);
        svg.appendChild(dot);
      });
    });

    if (opts.showXLabels !== false && data.length > 0) {
      var labelStep = Math.max(1, Math.floor(data.length / 8));
      data.forEach(function (d, i) {
        if (i % labelStep !== 0 && i !== data.length - 1) return;
        var txt = document.createElementNS(svgNS, 'text');
        txt.setAttribute('x', xPos(i));
        txt.setAttribute('y', h - 6);
        txt.setAttribute('text-anchor', 'middle');
        txt.setAttribute('fill', colors.muted);
        txt.setAttribute('font-size', '10');
        txt.textContent = String(d[xKey] || '');
        svg.appendChild(txt);
      });
    }

    if (opts.legend !== false && seriesKeys.length > 1) {
      var legend = document.createElement('div');
      legend.style.cssText = 'display:flex;flex-wrap:wrap;justify-content:center;gap:12px;margin-top:8px;';
      seriesKeys.forEach(function (key, si) {
        var color = seriesColors[key] || colors.palette[si % colors.palette.length];
        var item = document.createElement('span');
        item.style.cssText = 'display:flex;align-items:center;gap:6px;font-family:var(--fm,\'Inter\',sans-serif);font-size:11px;color:' + colors.text + ';';
        item.innerHTML = '<span style="width:10px;height:10px;border-radius:50%;background:' + color + ';display:inline-block;"></span>' + key;
        legend.appendChild(item);
      });
      container.appendChild(legend);
    }

    container.appendChild(svg);
  }

  function bar(container, data, opts) {
    if (typeof container === 'string') container = document.querySelector(container);
    if (!container || !data || !data.length) { renderEmpty(container); return; }
    opts = opts || {};
    var colors = getColors(opts);

    var svgNS = 'http://www.w3.org/2000/svg';
    var id = chartId();
    var w = opts.width || container.clientWidth || 400;
    var h = opts.height || 240;
    var pad = opts.padding || { top: 20, right: 10, bottom: 30, left: 45 };
    if (!opts.padding) { pad.top = 20; pad.right = 10; pad.bottom = 30; pad.left = 45; }

    var xKey = opts.xKey || 'label';
    var yKey = opts.yKey || 'value';
    var barColor = opts.color || colors.palette[0];
    var barRadius = opts.barRadius || 4;

    var plotW = w - pad.left - pad.right;
    var plotH = h - pad.top - pad.bottom;

    var vals = data.map(function (d) { return sanitizeVal(d[yKey]); });
    var maxVal = opts.maxY !== undefined ? opts.maxY : Math.max.apply(null, vals);
    if (maxVal <= 0) maxVal = 1;
    var minVal = 0;

    var barW = Math.min(plotW / data.length * 0.7, 48);
    var gap = (plotW - barW * data.length) / (data.length + 1);

    var svg = createSvg(w, h, id);

    if (opts.showGrid !== false) {
      var gridLines = 5;
      for (var gi = 0; gi <= gridLines; gi++) {
        var gy = pad.top + (plotH / gridLines) * gi;
        var lineEl = document.createElementNS(svgNS, 'line');
        lineEl.setAttribute('x1', pad.left);
        lineEl.setAttribute('y1', gy);
        lineEl.setAttribute('x2', w - pad.right);
        lineEl.setAttribute('y2', gy);
        lineEl.setAttribute('stroke', colors.grid);
        lineEl.setAttribute('stroke-width', '1');
        lineEl.setAttribute('stroke-dasharray', '4,4');
        svg.appendChild(lineEl);
        var gv = maxVal - (maxVal / gridLines) * gi;
        var txt = document.createElementNS(svgNS, 'text');
        txt.setAttribute('x', pad.left - 8);
        txt.setAttribute('y', gy + 4);
        txt.setAttribute('text-anchor', 'end');
        txt.setAttribute('fill', colors.muted);
        txt.setAttribute('font-size', '10');
        txt.textContent = formatNum(gv);
        svg.appendChild(txt);
      }
    }

    data.forEach(function (d, i) {
      var v = sanitizeVal(d[yKey]);
      var barH = (v / maxVal) * plotH;
      var x = pad.left + gap + i * (barW + gap);
      var y = pad.top + plotH - barH;

      var rect = document.createElementNS(svgNS, 'rect');
      rect.setAttribute('x', x);
      rect.setAttribute('y', y);
      rect.setAttribute('width', barW);
      rect.setAttribute('height', barH);
      rect.setAttribute('rx', barRadius);
      rect.setAttribute('ry', barRadius);
      rect.setAttribute('fill', opts.getColor ? opts.getColor(d, i) : barColor);
      rect.setAttribute('opacity', '0.85');
      rect.style.cursor = 'pointer';
      rect.addEventListener('mouseenter', function (e) {
        rect.setAttribute('opacity', '1');
        showTooltip(e, formatNum(v));
      });
      rect.addEventListener('mouseleave', function () {
        rect.setAttribute('opacity', '0.85');
        hideTooltip();
      });
      svg.appendChild(rect);

      if (opts.showValues === true && barH > 20) {
        var vt = document.createElementNS(svgNS, 'text');
        vt.setAttribute('x', x + barW / 2);
        vt.setAttribute('y', y - 6);
        vt.setAttribute('text-anchor', 'middle');
        vt.setAttribute('fill', colors.text);
        vt.setAttribute('font-size', '10');
        vt.textContent = formatNum(v);
        svg.appendChild(vt);
      }

      if (opts.showXLabels !== false) {
        var lb = document.createElementNS(svgNS, 'text');
        lb.setAttribute('x', x + barW / 2);
        lb.setAttribute('y', h - 6);
        lb.setAttribute('text-anchor', 'middle');
        lb.setAttribute('fill', colors.muted);
        lb.setAttribute('font-size', '10');
        lb.textContent = String(d[xKey] || '');
        svg.appendChild(lb);
      }
    });

    container.appendChild(svg);
  }

  function doughnut(container, data, opts) {
    if (typeof container === 'string') container = document.querySelector(container);
    if (!container || !data || !data.length) { renderEmpty(container); return; }
    opts = opts || {};
    var colors = getColors(opts);

    var svgNS = 'http://www.w3.org/2000/svg';
    var id = chartId();
    var w = opts.width || container.clientWidth || 240;
    var h = opts.height || w;
    var cx = w / 2;
    var cy = h / 2;
    var outerR = Math.min(cx, cy) - 20;
    var innerR = outerR * (opts.innerRatio || 0.6);
    var total = data.reduce(function (s, d) { return s + sanitizeVal(d.value); }, 0) || 1;

    var svg = createSvg(w, h, id);

    var angle = -Math.PI / 2;
    data.forEach(function (d, i) {
      var v = sanitizeVal(d.value);
      var sliceAngle = (v / total) * 2 * Math.PI;
      var color = d.color || colors.palette[i % colors.palette.length];
      var path = describeArc(cx, cy, outerR, angle, angle + sliceAngle) +
        describeArc(cx, cy, innerR, angle + sliceAngle, angle);
      var p = document.createElementNS(svgNS, 'path');
      p.setAttribute('d', path);
      p.setAttribute('fill', color);
      p.setAttribute('opacity', '0.9');
      p.style.cursor = 'pointer';
      p.addEventListener('mouseenter', function (e) {
        p.setAttribute('opacity', '1');
        var pct = ((v / total) * 100).toFixed(1);
        showTooltip(e, d.label + ': ' + formatNum(v) + ' (' + pct + '%)');
      });
      p.addEventListener('mouseleave', function () { p.setAttribute('opacity', '0.9'); hideTooltip(); });
      svg.appendChild(p);
      angle += sliceAngle;
    });

    if (opts.centerText) {
      var ct = document.createElementNS(svgNS, 'text');
      ct.setAttribute('x', cx);
      ct.setAttribute('y', cy);
      ct.setAttribute('text-anchor', 'middle');
      ct.setAttribute('dominant-baseline', 'central');
      ct.setAttribute('fill', colors.text);
      ct.setAttribute('font-size', '16');
      ct.setAttribute('font-weight', '600');
      ct.textContent = opts.centerText;
      svg.appendChild(ct);
    }

    container.appendChild(svg);

    if (opts.legend !== false) {
      var legend = document.createElement('div');
      legend.style.cssText = 'display:flex;flex-wrap:wrap;justify-content:center;gap:10px;margin-top:10px;';
      data.forEach(function (d, i) {
        var color = d.color || colors.palette[i % colors.palette.length];
        var item = document.createElement('span');
        item.style.cssText = 'display:flex;align-items:center;gap:6px;font-family:var(--fm,\'Inter\',sans-serif);font-size:11px;color:' + colors.text + ';';
        item.innerHTML = '<span style="width:10px;height:10px;border-radius:3px;background:' + color + ';display:inline-block;"></span>' + d.label;
        legend.appendChild(item);
      });
      container.appendChild(legend);
    }
  }

  function scoreRing(container, percent, opts) {
    if (typeof container === 'string') container = document.querySelector(container);
    if (!container) return;
    opts = opts || {};
    var pct = Math.min(100, Math.max(0, Number(percent) || 0));
    var w = opts.size || 120;
    var h = w;
    var stroke = opts.strokeWidth || 8;
    var r = (w / 2) - stroke;
    var cx = w / 2;
    var cy = h / 2;
    var circumference = 2 * Math.PI * r;
    var offset = circumference * (1 - pct / 100);

    var svgNS = 'http://www.w3.org/2000/svg';
    var svg = createSvg(w, h);

    var bgCircle = document.createElementNS(svgNS, 'circle');
    bgCircle.setAttribute('cx', cx);
    bgCircle.setAttribute('cy', cy);
    bgCircle.setAttribute('r', r);
    bgCircle.setAttribute('fill', 'none');
    bgCircle.setAttribute('stroke', gridColor);
    bgCircle.setAttribute('stroke-width', stroke);
    svg.appendChild(bgCircle);

    var fgCircle = document.createElementNS(svgNS, 'circle');
    fgCircle.setAttribute('cx', cx);
    fgCircle.setAttribute('cy', cy);
    fgCircle.setAttribute('r', r);
    fgCircle.setAttribute('fill', 'none');
    fgCircle.setAttribute('stroke', pct >= 80 ? accent : pct >= 50 ? warningColor : dangerColor);
    fgCircle.setAttribute('stroke-width', stroke);
    fgCircle.setAttribute('stroke-linecap', 'round');
    fgCircle.setAttribute('stroke-dasharray', circumference);
    fgCircle.setAttribute('stroke-dashoffset', offset);
    fgCircle.setAttribute('transform', 'rotate(-90 ' + cx + ' ' + cy + ')');
    fgCircle.style.transition = 'stroke-dashoffset 0.6s ease';
    svg.appendChild(fgCircle);

    var txt = document.createElementNS(svgNS, 'text');
    txt.setAttribute('x', cx);
    txt.setAttribute('y', cy - 2);
    txt.setAttribute('text-anchor', 'middle');
    txt.setAttribute('dominant-baseline', 'central');
    txt.setAttribute('fill', textColor);
    txt.setAttribute('font-size', '24');
    txt.setAttribute('font-weight', '700');
    txt.textContent = pct + '%';
    svg.appendChild(txt);

    if (opts.label) {
      var lbl = document.createElementNS(svgNS, 'text');
      lbl.setAttribute('x', cx);
      lbl.setAttribute('y', cy + 16);
      lbl.setAttribute('text-anchor', 'middle');
      lbl.setAttribute('fill', mutedColor);
      lbl.setAttribute('font-size', '10');
      lbl.textContent = opts.label;
      svg.appendChild(lbl);
    }

    container.appendChild(svg);
  }

  function gauge(container, percent, opts) {
    if (typeof container === 'string') container = document.querySelector(container);
    if (!container) return;
    opts = opts || {};
    var pct = Math.min(100, Math.max(0, Number(percent) || 0));
    var w = opts.width || container.clientWidth || 300;
    var h = opts.height || w * 0.5;
    var stroke = opts.strokeWidth || 12;
    var cx = w / 2;
    var r = Math.min(cx, h - 20) - stroke;
    var cy = h - 10;
    var startAngle = Math.PI * 0.75;
    var endAngle = Math.PI * 2.25;
    var totalAngle = endAngle - startAngle;

    var svgNS = 'http://www.w3.org/2000/svg';
    var svg = createSvg(w, h);

    var bgArc = describeArcCartesian(cx, cy, r, startAngle, endAngle);
    var bg = document.createElementNS(svgNS, 'path');
    bg.setAttribute('d', bgArc);
    bg.setAttribute('fill', 'none');
    bg.setAttribute('stroke', gridColor);
    bg.setAttribute('stroke-width', stroke);
    bg.setAttribute('stroke-linecap', 'round');
    svg.appendChild(bg);

    var valAngle = startAngle + (pct / 100) * totalAngle;
    var fgArc = describeArcCartesian(cx, cy, r, startAngle, valAngle);
    var fg = document.createElementNS(svgNS, 'path');
    fg.setAttribute('d', fgArc);
    fg.setAttribute('fill', 'none');
    fg.setAttribute('stroke', opts.color || (pct >= 80 ? accent : pct >= 50 ? warningColor : dangerColor));
    fg.setAttribute('stroke-width', stroke);
    fg.setAttribute('stroke-linecap', 'round');
    svg.appendChild(fg);

    var txt = document.createElementNS(svgNS, 'text');
    txt.setAttribute('x', cx);
    txt.setAttribute('y', cy - 24);
    txt.setAttribute('text-anchor', 'middle');
    txt.setAttribute('fill', textColor);
    txt.setAttribute('font-size', '28');
    txt.setAttribute('font-weight', '700');
    txt.textContent = Math.round(pct) + '%';
    svg.appendChild(txt);

    if (opts.label) {
      var lbl = document.createElementNS(svgNS, 'text');
      lbl.setAttribute('x', cx);
      lbl.setAttribute('y', cy - 4);
      lbl.setAttribute('text-anchor', 'middle');
      lbl.setAttribute('fill', mutedColor);
      lbl.setAttribute('font-size', '11');
      lbl.textContent = opts.label;
      svg.appendChild(lbl);
    }

    container.appendChild(svg);
  }

  function createSvg(w, h, id) {
    var svgNS = 'http://www.w3.org/2000/svg';
    var svg = document.createElementNS(svgNS, 'svg');
    svg.setAttribute('width', w);
    svg.setAttribute('height', h);
    svg.setAttribute('viewBox', '0 0 ' + w + ' ' + h);
    svg.style.cssText = 'max-width:100%;display:block;';
    if (id) svg.setAttribute('id', id);
    return svg;
  }

  function describeArc(cx, cy, r, startAngle, endAngle, reverse) {
    var start = polarToCartesian(cx, cy, r, endAngle);
    var end = polarToCartesian(cx, cy, r, startAngle);
    var largeArcFlag = endAngle - startAngle <= Math.PI ? '0' : '1';
    if (reverse) {
      return ' L ' + start.x + ' ' + start.y + ' A ' + r + ' ' + r + ' 0 ' + largeArcFlag + ' 1 ' + end.x + ' ' + end.y;
    }
    return ' M ' + start.x + ' ' + start.y + ' A ' + r + ' ' + r + ' 0 ' + largeArcFlag + ' 1 ' + end.x + ' ' + end.y;
  }

  function describeArcCartesian(cx, cy, r, startAngle, endAngle) {
    var s = polarToCartesian(cx, cy, r, startAngle);
    var e = polarToCartesian(cx, cy, r, endAngle);
    var large = endAngle - startAngle <= Math.PI ? '0' : '1';
    return 'M ' + s.x + ' ' + s.y + ' A ' + r + ' ' + r + ' 0 ' + large + ' 1 ' + e.x + ' ' + e.y;
  }

  function polarToCartesian(cx, cy, r, angle) {
    return { x: cx + r * Math.cos(angle), y: cy + r * Math.sin(angle) };
  }

  function buildSmoothPath(points, tension) {
    if (points.length < 4) return 'M ' + points[0] + ' L ' + points.slice(1).join(' L ');
    tension = tension || 0.3;
    var parts = [];
    var pts = points.map(function (p) { var sp = p.split(','); return { x: parseFloat(sp[0]), y: parseFloat(sp[1]) }; });
    parts.push('M ' + pts[0].x + ' ' + pts[0].y);
    for (var i = 0; i < pts.length - 1; i++) {
      var p0 = pts[Math.max(0, i - 1)];
      var p1 = pts[i];
      var p2 = pts[i + 1];
      var p3 = pts[Math.min(i + 2, pts.length - 1)];
      var cp1x = p1.x + (p2.x - p0.x) * tension;
      var cp1y = p1.y + (p2.y - p0.y) * tension;
      var cp2x = p2.x - (p3.x - p1.x) * tension;
      var cp2y = p2.y - (p3.y - p1.y) * tension;
      parts.push('C ' + cp1x + ' ' + cp1y + ', ' + cp2x + ' ' + cp2y + ', ' + p2.x + ' ' + p2.y);
    }
    return parts.join(' ');
  }

  var tooltipEl = null;
  function showTooltip(e, html) {
    if (!tooltipEl || !document.body.contains(tooltipEl)) {
      tooltipEl = document.createElement('div');
      tooltipEl.className = 'angazia-chart-tooltip';
      tooltipEl.style.cssText = 'position:fixed;z-index:99999;background:#1f2937;color:#fff;padding:6px 12px;border-radius:8px;font-family:var(--fm,\'Inter\',sans-serif);font-size:12px;font-weight:500;pointer-events:none;white-space:nowrap;box-shadow:0 4px 12px rgba(0,0,0,0.2);max-width:300px;white-space:normal;';
      document.body.appendChild(tooltipEl);
    }
    tooltipEl.innerHTML = html;
    tooltipEl.style.display = '';
    var ex = e.clientX || e.pageX || 0;
    var ey = e.clientY || e.pageY || 0;
    var tw = tooltipEl.offsetWidth;
    var th = tooltipEl.offsetHeight;
    var l = ex + 12;
    var t = ey - th - 10;
    if (l + tw > window.innerWidth) l = ex - tw - 12;
    if (t < 8) t = ey + 12;
    tooltipEl.style.left = l + 'px';
    tooltipEl.style.top = t + 'px';
  }

  function hideTooltip() {
    if (tooltipEl) tooltipEl.style.display = 'none';
  }

  function formatNum(n) {
    if (n >= 1000000) return (n / 1000000).toFixed(1) + 'M';
    if (n >= 1000) return (n / 1000).toFixed(1) + 'K';
    return Number(n.toFixed(1)).toString();
  }

  function renderEmpty(container) {
    if (!container) return;
    var el = document.createElement('div');
    el.style.cssText = 'display:flex;align-items:center;justify-content:center;height:120px;color:' + mutedColor + ';font-family:var(--fm,\'Inter\',sans-serif);font-size:13px;';
    el.textContent = 'No data available';
    container.appendChild(el);
  }

  return {
    line: line,
    bar: bar,
    doughnut: doughnut,
    scoreRing: scoreRing,
    gauge: gauge,
  };
})();
