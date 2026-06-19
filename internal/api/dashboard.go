package api

import "net/http"

func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(dashboardHTML))
}

const dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>lbsim — Control Plane</title>
<style>
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { font-family: 'SF Mono', 'Fira Code', monospace; background: #0d1117; color: #e6edf3; min-height: 100vh; padding: 24px; }
  h1 { font-size: 18px; font-weight: 600; color: #58a6ff; margin-bottom: 4px; }
  .subtitle { font-size: 12px; color: #8b949e; margin-bottom: 24px; }
  .panel { background: #161b22; border: 1px solid #30363d; border-radius: 8px; padding: 18px; margin-bottom: 18px; }
  .panel h2 { font-size: 11px; color: #8b949e; margin-bottom: 14px; font-weight: 500; text-transform: uppercase; letter-spacing: .08em; }
  .cluster { background: #161b22; border: 1px solid #30363d; border-radius: 8px; margin-bottom: 18px; overflow: hidden; }
  .cluster-header { padding: 12px 16px; border-bottom: 1px solid #30363d; display: flex; align-items: center; gap: 10px; }
  .cluster-name { font-size: 14px; font-weight: 600; color: #f0f6fc; }
  .cluster-count { font-size: 12px; color: #8b949e; }
  .backends { display: grid; grid-template-columns: repeat(auto-fill, minmax(270px, 1fr)); gap: 10px; padding: 14px; }
  .backend { background: #0d1117; border: 1px solid #30363d; border-radius: 6px; padding: 12px; position: relative; transition: border-color .2s; }
  .backend.last-picked { border-color: #58a6ff; }
  .backend-id { font-size: 12px; font-weight: 600; color: #f0f6fc; margin-bottom: 6px; }
  .backend-meta { font-size: 11px; color: #8b949e; line-height: 1.8; margin-bottom: 8px; }
  .dispatch-count { position: absolute; top: 10px; right: 10px; font-size: 11px; color: #58a6ff; background: #0d2040; border: 1px solid #1f4070; border-radius: 10px; padding: 1px 7px; }
  .dist-bar-wrap { margin: 6px 0 10px; height: 6px; background: #21262d; border-radius: 3px; overflow: hidden; }
  .dist-bar-fill { height: 100%; border-radius: 3px; transition: width .4s ease; background: #1f6feb; }
  .dist-pct { font-size: 10px; color: #58a6ff; margin-bottom: 8px; }
  .badge { display: inline-block; padding: 2px 7px; border-radius: 10px; font-size: 10px; font-weight: 600; margin-bottom: 8px; }
  .badge-healthy   { background: #1a3a2a; color: #3fb950; border: 1px solid #238636; }
  .badge-unhealthy { background: #3a1a1a; color: #f85149; border: 1px solid #da3633; }
  .badge-draining  { background: #3a2e1a; color: #d29922; border: 1px solid #9e6a03; }
  .actions { display: flex; gap: 5px; flex-wrap: wrap; }
  .btn { padding: 3px 9px; font-size: 11px; font-family: inherit; border-radius: 4px; border: 1px solid; cursor: pointer; font-weight: 500; }
  .btn:hover { opacity: .8; }
  .btn-green  { background: #1a3a2a; color: #3fb950; border-color: #238636; }
  .btn-red    { background: #3a1a1a; color: #f85149; border-color: #da3633; }
  .btn-yellow { background: #3a2e1a; color: #d29922; border-color: #9e6a03; }
  .btn-gray   { background: #21262d; color: #8b949e; border-color: #30363d; }
  .btn-blue   { background: #0d2040; color: #58a6ff; border-color: #1f4070; }
  .btn-stop   { background: #3a1a1a; color: #f85149; border-color: #da3633; }
  .form-row { display: flex; gap: 8px; flex-wrap: wrap; align-items: flex-end; }
  .field { display: flex; flex-direction: column; gap: 4px; }
  .field label { font-size: 11px; color: #8b949e; }
  .field input, .field select { background: #0d1117; border: 1px solid #30363d; color: #e6edf3; padding: 5px 9px; border-radius: 4px; font-family: inherit; font-size: 12px; }
  .field input { width: 140px; }
  .field input.short { width: 80px; }
  .field input:focus { outline: none; border-color: #58a6ff; }
  .algo-group { display: flex; gap: 5px; }
  .algo-btn { padding: 5px 10px; font-size: 11px; font-family: inherit; border-radius: 4px; border: 1px solid #30363d; cursor: pointer; background: #21262d; color: #8b949e; }
  .algo-btn.active { background: #0d2040; color: #58a6ff; border-color: #1f4070; }
  .result-box { margin-top: 12px; padding: 10px 12px; background: #0d1117; border: 1px solid #30363d; border-radius: 6px; font-size: 12px; min-height: 38px; color: #8b949e; }
  /* dispatcher panel */
  .disp-stats { display: flex; gap: 24px; margin-bottom: 14px; }
  .disp-stat { display: flex; flex-direction: column; gap: 2px; }
  .disp-stat-val { font-size: 22px; font-weight: 700; color: #f0f6fc; }
  .disp-stat-lbl { font-size: 10px; color: #8b949e; text-transform: uppercase; letter-spacing: .06em; }
  .disp-bars { display: flex; flex-direction: column; gap: 8px; margin-top: 14px; }
  .disp-row { display: grid; grid-template-columns: 130px 1fr 60px; gap: 8px; align-items: center; font-size: 11px; }
  .disp-row-id { color: #e6edf3; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .disp-bar-bg { background: #21262d; border-radius: 3px; height: 8px; overflow: hidden; }
  .disp-bar-fg { height: 100%; border-radius: 3px; transition: width .3s ease; }
  .disp-bar-dc-tokyo  { background: #1f6feb; }
  .disp-bar-dc-osaka  { background: #a371f7; }
  .disp-bar-dc-other  { background: #3fb950; }
  .disp-row-count { color: #8b949e; text-align: right; }
  .running-dot { display: inline-block; width: 7px; height: 7px; border-radius: 50%; background: #3fb950; box-shadow: 0 0 6px #3fb950; margin-right: 6px; animation: blink 1.2s infinite; }
  @keyframes blink { 0%,100%{opacity:1} 50%{opacity:.3} }
  .stopped-dot { display: inline-block; width: 7px; height: 7px; border-radius: 50%; background: #8b949e; margin-right: 6px; }
  .pulse { width: 7px; height: 7px; border-radius: 50%; display: inline-block; margin-right: 5px; }
  .pulse-green  { background: #3fb950; box-shadow: 0 0 5px #3fb950; }
  .pulse-red    { background: #f85149; }
  .pulse-yellow { background: #d29922; }
  .ticker { font-size: 11px; color: #8b949e; margin-top: 6px; }
  .empty { color: #8b949e; font-size: 12px; padding: 18px 14px; }
  .error-bar { background: #3a1a1a; border: 1px solid #da3633; color: #f85149; padding: 7px 12px; border-radius: 6px; font-size: 12px; margin-bottom: 14px; display: none; }
  .sep { border: none; border-top: 1px solid #30363d; margin: 14px 0; }
</style>
</head>
<body>
<h1>lbsim — Control Plane</h1>
<p class="subtitle">L1 registry · L2 health checking · L3 routing · L4 dispatcher · auto-refreshes every 2 s</p>

<div id="error-bar" class="error-bar"></div>

<!-- ── Layer 4: Dispatcher ───────────────────────────── -->
<div class="panel">
  <h2><span id="disp-dot" class="stopped-dot"></span>Live Dispatcher — Layer 4</h2>
  <div class="form-row">
    <div class="field">
      <label>cluster</label>
      <input id="d-cluster" value="api" class="short">
    </div>
    <div class="field">
      <label>algorithm</label>
      <div class="algo-group" id="d-algo-group">
        <button class="algo-btn active" onclick="setDispAlgo('round-robin',this)">round-robin</button>
        <button class="algo-btn" onclick="setDispAlgo('weighted',this)">weighted</button>
        <button class="algo-btn" onclick="setDispAlgo('consistent-hash',this)">consistent-hash</button>
      </div>
    </div>
    <div class="field">
      <label>req/s</label>
      <input id="d-rps" value="5" type="number" min="1" max="50" class="short">
    </div>
    <div class="field"><label>&nbsp;</label><button id="d-start-btn" class="btn btn-blue" onclick="dispStart()">▶ Start</button></div>
    <div class="field"><label>&nbsp;</label><button class="btn btn-stop" onclick="dispStop()">■ Stop</button></div>
    <div class="field"><label>&nbsp;</label><button class="btn btn-gray" onclick="dispReset()">↺ Reset</button></div>
  </div>

  <div class="disp-stats" style="margin-top:16px">
    <div class="disp-stat"><span class="disp-stat-val" id="d-total">0</span><span class="disp-stat-lbl">total requests</span></div>
    <div class="disp-stat"><span class="disp-stat-val" id="d-rps-live">0.0</span><span class="disp-stat-lbl">req/s (10s avg)</span></div>
    <div class="disp-stat"><span class="disp-stat-val" id="d-algo-live">—</span><span class="disp-stat-lbl">algorithm</span></div>
  </div>

  <div id="disp-bars" class="disp-bars"></div>
</div>

<!-- ── Layer 3: Manual routing ──────────────────────── -->
<div class="panel">
  <h2>Manual Route — Layer 3</h2>
  <div class="form-row">
    <div class="field"><label>cluster</label><input id="r-cluster" value="api" class="short"></div>
    <div class="field">
      <label>algorithm</label>
      <div class="algo-group">
        <button class="algo-btn active" onclick="setRouteAlgo('round-robin',this)">round-robin</button>
        <button class="algo-btn" onclick="setRouteAlgo('weighted',this)">weighted</button>
        <button class="algo-btn" onclick="setRouteAlgo('consistent-hash',this)">consistent-hash</button>
      </div>
    </div>
    <div class="field"><label>key</label><input id="r-key" placeholder="user-123" class="short"></div>
    <div class="field"><label>&nbsp;</label><button class="btn btn-blue" onclick="manualDispatch(1)">Dispatch</button></div>
    <div class="field"><label>&nbsp;</label><button class="btn btn-gray" onclick="manualDispatch(20)">× 20</button></div>
  </div>
  <div class="result-box" id="result-box">pick an algorithm and click Dispatch</div>
</div>

<!-- ── Layer 1: Register backend ────────────────────── -->
<div class="panel">
  <h2>Register Backend — Layer 1</h2>
  <div class="form-row">
    <div class="field"><label>cluster</label><input id="f-cluster" value="api"></div>
    <div class="field"><label>id</label><input id="f-id" placeholder="singapore-api-01"></div>
    <div class="field"><label>addr</label><input id="f-addr" placeholder="10.0.3.1:8080"></div>
    <div class="field"><label>dc</label><input id="f-dc" placeholder="singapore" class="short"></div>
    <div class="field"><label>weight</label><input id="f-weight" placeholder="3" class="short" type="number"></div>
    <div class="field"><label>&nbsp;</label><button class="btn btn-green" onclick="addBackend()">+ Register</button></div>
  </div>
</div>

<!-- ── Cluster state ────────────────────────────────── -->
<div id="clusters"></div>
<p class="ticker" id="ticker">refreshing...</p>

<script>
let routeAlgo = 'round-robin';
let dispAlgo  = 'round-robin';
let dispStats = {};
let routeStats = {};
let lastPicked = null;
let lastClusters = [];

// dc → bar colour class
function dcClass(dc) {
  if (dc === 'tokyo') return 'disp-bar-dc-tokyo';
  if (dc === 'osaka') return 'disp-bar-dc-osaka';
  return 'disp-bar-dc-other';
}

// ── dispatcher controls ──────────────────────────────
function setDispAlgo(algo, el) {
  dispAlgo = algo;
  document.querySelectorAll('#d-algo-group .algo-btn').forEach(b => b.classList.remove('active'));
  el.classList.add('active');
}

async function dispStart() {
  const cluster = document.getElementById('d-cluster').value.trim() || 'api';
  const rps     = parseInt(document.getElementById('d-rps').value) || 5;
  try {
    const res = await fetch('/dispatcher/start', {
      method: 'POST', headers: {'Content-Type':'application/json'},
      body: JSON.stringify({cluster, algo: dispAlgo, rps}),
    });
    if (!res.ok) { const e = await res.json(); throw new Error(e.error); }
  } catch(e) { showError(e.message); }
}

async function dispStop() {
  await fetch('/dispatcher/stop', {method:'POST'});
}

async function dispReset() {
  await fetch('/dispatcher/reset', {method:'POST'});
  dispStats = {};
  renderDispBars({backends:[], total:0, rps:0, algo:'—', running:false});
}

async function loadDispStats() {
  try {
    const res = await fetch('/dispatcher/stats');
    dispStats = await res.json();
    renderDispBars(dispStats);
  } catch(_) {}
}

function renderDispBars(s) {
  document.getElementById('d-total').textContent    = (s.total || 0).toLocaleString();
  document.getElementById('d-rps-live').textContent = (s.rps || 0).toFixed(1);
  document.getElementById('d-algo-live').textContent = s.algo || '—';

  const dot = document.getElementById('disp-dot');
  dot.className = s.running ? 'running-dot' : 'stopped-dot';

  const backends = s.backends || [];
  const total = s.total || 0;
  const wrap = document.getElementById('disp-bars');
  if (backends.length === 0) { wrap.innerHTML = ''; return; }

  // sort by total desc
  backends.sort((a,b) => b.total - a.total);

  wrap.innerHTML = backends.map(b => {
    const pct = total > 0 ? ((b.total / total) * 100).toFixed(1) : '0.0';
    const cls = dcClass(b.dc);
    return ` + "`" + `
    <div class="disp-row">
      <div class="disp-row-id" title="${b.id}">${b.id}</div>
      <div class="disp-bar-bg"><div class="disp-bar-fg ${cls}" style="width:${pct}%"></div></div>
      <div class="disp-row-count">${b.total.toLocaleString()} <span style="color:#484f58">(${pct}%)</span></div>
    </div>` + "`" + `;
  }).join('');
}

// ── manual routing ───────────────────────────────────
function setRouteAlgo(algo, el) {
  routeAlgo = algo;
  el.closest('.algo-group').querySelectorAll('.algo-btn').forEach(b => b.classList.remove('active'));
  el.classList.add('active');
}

async function manualDispatch(times) {
  const cluster = document.getElementById('r-cluster').value.trim() || 'api';
  const key     = document.getElementById('r-key').value.trim() || 'default';
  let last;
  for (let i = 0; i < times; i++) {
    try {
      const res = await fetch('/route', {
        method:'POST', headers:{'Content-Type':'application/json'},
        body: JSON.stringify({cluster, algo: routeAlgo, key}),
      });
      if (!res.ok) { const e = await res.json(); throw new Error(e.error); }
      last = await res.json();
    } catch(e) { showError(e.message); return; }
  }
  if (last) {
    lastPicked = last.backend.id;
    const b = last.backend;
    const sfx = times > 1 ? ` + "`" + ` (after ${times} dispatches)` + "`" + ` : '';
    document.getElementById('result-box').innerHTML =
      ` + "`" + `→ <strong style="color:#3fb950">${b.id}</strong> &nbsp;dc:${b.dc} weight:${b.weight} algo:${last.algo}${sfx}` + "`" + `;
    await loadRouteStats();
    render(lastClusters);
  }
}

async function loadRouteStats() {
  try {
    const res = await fetch('/route/stats');
    routeStats = await res.json();
  } catch(_) {}
}

// ── backend registration ─────────────────────────────
async function addBackend() {
  const cluster = document.getElementById('f-cluster').value.trim();
  const id      = document.getElementById('f-id').value.trim();
  const addr    = document.getElementById('f-addr').value.trim();
  const dc      = document.getElementById('f-dc').value.trim();
  const weight  = parseInt(document.getElementById('f-weight').value) || 1;
  if (!cluster || !id || !addr) { showError('cluster, id and addr required'); return; }
  try {
    const res = await fetch(` + "`" + `/clusters/${cluster}/backends` + "`" + `, {
      method:'POST', headers:{'Content-Type':'application/json'},
      body: JSON.stringify({id, addr, dc, weight}),
    });
    if (!res.ok) throw new Error((await res.json()).error);
    document.getElementById('f-id').value = '';
    document.getElementById('f-addr').value = '';
    load();
  } catch(e) { showError(e.message); }
}

// ── cluster / backend rendering ──────────────────────
async function load() {
  try {
    const [cRes] = await Promise.all([fetch('/clusters'), loadRouteStats(), loadDispStats()]);
    if (!cRes.ok) throw new Error(await cRes.text());
    lastClusters = await cRes.json();
    render(lastClusters);
    document.getElementById('error-bar').style.display = 'none';
    document.getElementById('ticker').textContent = 'last updated: ' + new Date().toLocaleTimeString();
  } catch(e) { showError('fetch failed: ' + e.message); }
}

function render(clusters) {
  const el = document.getElementById('clusters');
  if (!clusters || clusters.length === 0) { el.innerHTML = '<p class="empty">no clusters</p>'; return; }
  el.innerHTML = clusters.map(c => {
    const backends = Object.values(c.backends || {});
    const healthy = backends.filter(b => b.health === 'healthy').length;
    return ` + "`" + `
    <div class="cluster">
      <div class="cluster-header">
        <span class="cluster-name">${c.name}</span>
        <span class="cluster-count">${healthy}/${backends.length} healthy</span>
      </div>
      <div class="backends">
        ${backends.length === 0 ? '<p class="empty">no backends</p>' : backends.map(b => backendCard(c.name, b)).join('')}
      </div>
    </div>` + "`" + `;
  }).join('');
}

function backendCard(cluster, b) {
  const pulseClass = {healthy:'pulse-green',unhealthy:'pulse-red',draining:'pulse-yellow'}[b.health]||'pulse-red';
  const badgeCls   = {healthy:'badge-healthy',unhealthy:'badge-unhealthy',draining:'badge-draining'}[b.health]||'';
  const rCount     = routeStats[b.id] || 0;
  const dStat      = (dispStats.backends||[]).find(x => x.id === b.id);
  const dTotal     = dStat ? dStat.total : 0;
  const grandTotal = dispStats.total || 0;
  const pct        = grandTotal > 0 ? ((dTotal / grandTotal)*100).toFixed(1) : null;
  const isLast     = b.id === lastPicked;
  const lastSeen   = new Date(b.last_seen).toLocaleTimeString();
  return ` + "`" + `
  <div class="backend${isLast ? ' last-picked' : ''}">
    ${rCount > 0 ? ` + "`" + `<span class="dispatch-count">${rCount} req</span>` + "`" + ` : ''}
    <div class="backend-id"><span class="pulse ${pulseClass}"></span>${b.id}</div>
    <div class="backend-meta">addr &nbsp;${b.addr}<br>dc &nbsp;&nbsp;&nbsp;${b.dc}<br>weight ${b.weight}<br>seen &nbsp;${lastSeen}</div>
    ${pct !== null ? ` + "`" + `
      <div class="dist-bar-wrap"><div class="dist-bar-fill" style="width:${pct}%"></div></div>
      <div class="dist-pct">${dTotal.toLocaleString()} dispatched &mdash; ${pct}% of traffic</div>
    ` + "`" + ` : ''}
    <div><span class="badge ${badgeCls}">${b.health}</span></div>
    <div class="actions">
      ${b.health!=='healthy'   ? ` + "`" + `<button class="btn btn-green"  onclick="setHealth('${cluster}','${b.id}','healthy')">healthy</button>` + "`" + ` : ''}
      ${b.health!=='unhealthy' ? ` + "`" + `<button class="btn btn-red"    onclick="setHealth('${cluster}','${b.id}','unhealthy')">unhealthy</button>` + "`" + ` : ''}
      ${b.health!=='draining'  ? ` + "`" + `<button class="btn btn-yellow" onclick="setHealth('${cluster}','${b.id}','draining')">drain</button>` + "`" + ` : ''}
      <button class="btn btn-gray" onclick="removeBackend('${cluster}','${b.id}')">remove</button>
    </div>
  </div>` + "`" + `;
}

async function setHealth(cluster, id, health) {
  try {
    const res = await fetch(` + "`" + `/clusters/${cluster}/backends/${id}/health` + "`" + `, {
      method:'PATCH', headers:{'Content-Type':'application/json'}, body:JSON.stringify({health}),
    });
    if (!res.ok) throw new Error(await res.text());
    load();
  } catch(e) { showError(e.message); }
}

async function removeBackend(cluster, id) {
  await fetch(` + "`" + `/clusters/${cluster}/backends/${id}` + "`" + `, {method:'DELETE'});
  load();
}

function showError(msg) {
  const el = document.getElementById('error-bar');
  el.textContent = msg;
  el.style.display = 'block';
  setTimeout(() => el.style.display='none', 4000);
}

load();
setInterval(load, 2000);
// poll dispatcher stats more frequently while it may be running
setInterval(loadDispStats, 500);
</script>
</body>
</html>`
