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
<title>lbsim — Control Plane Dashboard</title>
<style>
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { font-family: 'SF Mono', 'Fira Code', monospace; background: #0d1117; color: #e6edf3; min-height: 100vh; padding: 24px; }
  h1 { font-size: 18px; font-weight: 600; color: #58a6ff; margin-bottom: 4px; }
  .subtitle { font-size: 12px; color: #8b949e; margin-bottom: 28px; }
  .panel { background: #161b22; border: 1px solid #30363d; border-radius: 8px; padding: 18px; margin-bottom: 20px; }
  .panel h2 { font-size: 12px; color: #8b949e; margin-bottom: 14px; font-weight: 500; text-transform: uppercase; letter-spacing: .05em; }
  .cluster { background: #161b22; border: 1px solid #30363d; border-radius: 8px; margin-bottom: 20px; overflow: hidden; }
  .cluster-header { padding: 14px 18px; border-bottom: 1px solid #30363d; display: flex; align-items: center; gap: 10px; }
  .cluster-name { font-size: 15px; font-weight: 600; color: #f0f6fc; }
  .cluster-count { font-size: 12px; color: #8b949e; }
  .backends { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 12px; padding: 16px; }
  .backend { background: #0d1117; border: 1px solid #30363d; border-radius: 6px; padding: 14px; position: relative; }
  .backend.last-picked { border-color: #58a6ff; box-shadow: 0 0 0 1px #58a6ff33; }
  .backend-id { font-size: 13px; font-weight: 600; color: #f0f6fc; margin-bottom: 8px; }
  .backend-meta { font-size: 11px; color: #8b949e; line-height: 1.9; margin-bottom: 10px; }
  .dispatch-count { position: absolute; top: 12px; right: 12px; font-size: 11px; color: #58a6ff; background: #0d2040; border: 1px solid #1f4070; border-radius: 10px; padding: 1px 7px; }
  .badge { display: inline-block; padding: 2px 8px; border-radius: 12px; font-size: 11px; font-weight: 600; margin-bottom: 10px; }
  .badge-healthy   { background: #1a3a2a; color: #3fb950; border: 1px solid #238636; }
  .badge-unhealthy { background: #3a1a1a; color: #f85149; border: 1px solid #da3633; }
  .badge-draining  { background: #3a2e1a; color: #d29922; border: 1px solid #9e6a03; }
  .actions { display: flex; gap: 6px; flex-wrap: wrap; }
  .btn { padding: 4px 10px; font-size: 11px; font-family: inherit; border-radius: 4px; border: 1px solid; cursor: pointer; font-weight: 500; transition: opacity 0.1s; }
  .btn:hover { opacity: 0.8; }
  .btn-green  { background: #1a3a2a; color: #3fb950; border-color: #238636; }
  .btn-red    { background: #3a1a1a; color: #f85149; border-color: #da3633; }
  .btn-yellow { background: #3a2e1a; color: #d29922; border-color: #9e6a03; }
  .btn-gray   { background: #21262d; color: #8b949e; border-color: #30363d; }
  .btn-blue   { background: #0d2040; color: #58a6ff; border-color: #1f4070; }
  .form-row { display: flex; gap: 8px; flex-wrap: wrap; align-items: flex-end; }
  .field { display: flex; flex-direction: column; gap: 4px; }
  .field label { font-size: 11px; color: #8b949e; }
  .field input, .field select { background: #0d1117; border: 1px solid #30363d; color: #e6edf3; padding: 6px 10px; border-radius: 4px; font-family: inherit; font-size: 12px; }
  .field input { width: 160px; }
  .field input.short { width: 90px; }
  .field input:focus, .field select:focus { outline: none; border-color: #58a6ff; }
  .algo-group { display: flex; gap: 6px; }
  .algo-btn { padding: 5px 12px; font-size: 11px; font-family: inherit; border-radius: 4px; border: 1px solid #30363d; cursor: pointer; background: #21262d; color: #8b949e; }
  .algo-btn.active { background: #0d2040; color: #58a6ff; border-color: #1f4070; }
  .result-box { margin-top: 14px; padding: 12px 14px; background: #0d1117; border: 1px solid #30363d; border-radius: 6px; font-size: 12px; min-height: 42px; color: #8b949e; }
  .result-box .picked { color: #3fb950; font-weight: 600; }
  .result-box .detail { color: #8b949e; }
  .pulse { width: 8px; height: 8px; border-radius: 50%; display: inline-block; margin-right: 6px; }
  .pulse-green  { background: #3fb950; box-shadow: 0 0 6px #3fb950; }
  .pulse-red    { background: #f85149; }
  .pulse-yellow { background: #d29922; }
  .ticker { font-size: 11px; color: #8b949e; margin-top: 8px; }
  .empty { color: #8b949e; font-size: 12px; padding: 20px 16px; }
  .error-bar { background: #3a1a1a; border: 1px solid #da3633; color: #f85149; padding: 8px 14px; border-radius: 6px; font-size: 12px; margin-bottom: 16px; display: none; }
</style>
</head>
<body>
<h1>lbsim — Control Plane</h1>
<p class="subtitle">Layer 1 registry · Layer 2 health checking · Layer 3 routing · auto-refreshes every 2 s</p>

<div id="error-bar" class="error-bar"></div>

<!-- Register backend -->
<div class="panel">
  <h2>Register backend</h2>
  <div class="form-row">
    <div class="field"><label>cluster</label><input id="f-cluster" value="api"></div>
    <div class="field"><label>id</label><input id="f-id" placeholder="singapore-api-01"></div>
    <div class="field"><label>addr</label><input id="f-addr" placeholder="10.0.3.1:8080"></div>
    <div class="field"><label>dc</label><input id="f-dc" placeholder="singapore" class="short"></div>
    <div class="field"><label>weight</label><input id="f-weight" placeholder="3" class="short" type="number"></div>
    <div class="field"><label>&nbsp;</label><button class="btn btn-green" onclick="addBackend()">+ Register</button></div>
  </div>
</div>

<!-- Route a request -->
<div class="panel">
  <h2>Route a request &mdash; Layer 3</h2>
  <div class="form-row">
    <div class="field">
      <label>cluster</label>
      <input id="r-cluster" value="api" class="short">
    </div>
    <div class="field">
      <label>algorithm</label>
      <div class="algo-group">
        <button class="algo-btn active" onclick="setAlgo('round-robin', this)">round-robin</button>
        <button class="algo-btn" onclick="setAlgo('weighted', this)">weighted</button>
        <button class="algo-btn" onclick="setAlgo('consistent-hash', this)">consistent-hash</button>
      </div>
    </div>
    <div class="field">
      <label>key (for consistent-hash)</label>
      <input id="r-key" placeholder="user-123" class="short">
    </div>
    <div class="field"><label>&nbsp;</label><button class="btn btn-blue" onclick="dispatch(1)">Dispatch</button></div>
    <div class="field"><label>&nbsp;</label><button class="btn btn-gray" onclick="dispatch(20)">× 20</button></div>
    <div class="field"><label>&nbsp;</label><button class="btn btn-gray" onclick="resetStats()">reset stats</button></div>
  </div>
  <div class="result-box" id="result-box">pick an algorithm and click Dispatch to route a request</div>
</div>

<div id="clusters"></div>
<p class="ticker" id="ticker">refreshing...</p>

<script>
let currentAlgo = 'round-robin';
let dispatchStats = {};
let lastPicked = null;

function setAlgo(algo, el) {
  currentAlgo = algo;
  document.querySelectorAll('.algo-btn').forEach(b => b.classList.remove('active'));
  el.classList.add('active');
}

async function dispatch(times) {
  const cluster = document.getElementById('r-cluster').value.trim() || 'api';
  const key     = document.getElementById('r-key').value.trim() || 'default';
  let last;
  for (let i = 0; i < times; i++) {
    try {
      const res = await fetch('/route', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({cluster, algo: currentAlgo, key}),
      });
      if (!res.ok) { const e = await res.json(); throw new Error(e.error); }
      last = await res.json();
    } catch(e) { showError(e.message); return; }
  }
  if (last) {
    lastPicked = last.backend.id;
    const b = last.backend;
    const box = document.getElementById('result-box');
    const suffix = times > 1 ? ` + "`" + ` (after ${times} dispatches)` + "`" + ` : '';
    box.innerHTML = ` + "`" + `→ <span class="picked">${b.id}</span> <span class="detail">dc:${b.dc} weight:${b.weight} algo:${last.algo}${suffix}</span>` + "`" + `;
    await loadStats();
    render(lastClusters);
  }
}

async function resetStats() {
  dispatchStats = {};
  lastPicked = null;
  document.getElementById('result-box').textContent = 'stats reset';
  render(lastClusters);
}

async function loadStats() {
  try {
    const res = await fetch('/route/stats');
    dispatchStats = await res.json();
  } catch(_) {}
}

let lastClusters = [];

async function load() {
  try {
    const [cRes] = await Promise.all([fetch('/clusters'), loadStats()]);
    if (!cRes.ok) throw new Error(await cRes.text());
    lastClusters = await cRes.json();
    render(lastClusters);
    document.getElementById('error-bar').style.display = 'none';
    document.getElementById('ticker').textContent = 'last updated: ' + new Date().toLocaleTimeString();
  } catch(e) { showError('fetch failed: ' + e.message); }
}

function render(clusters) {
  const el = document.getElementById('clusters');
  if (!clusters || clusters.length === 0) {
    el.innerHTML = '<p class="empty">no clusters registered</p>';
    return;
  }
  el.innerHTML = clusters.map(c => {
    const backends = Object.values(c.backends || {});
    const healthyCount = backends.filter(b => b.health === 'healthy').length;
    return ` + "`" + `
    <div class="cluster">
      <div class="cluster-header">
        <span class="cluster-name">${c.name}</span>
        <span class="cluster-count">${healthyCount}/${backends.length} healthy</span>
      </div>
      <div class="backends">
        ${backends.length === 0 ? '<p class="empty">no backends</p>' : backends.map(b => backendCard(c.name, b)).join('')}
      </div>
    </div>` + "`" + `;
  }).join('');
}

function backendCard(cluster, b) {
  const pulseClass  = {healthy:'pulse-green', unhealthy:'pulse-red', draining:'pulse-yellow'}[b.health] || 'pulse-red';
  const badgeClass  = {healthy:'badge-healthy', unhealthy:'badge-unhealthy', draining:'badge-draining'}[b.health] || '';
  const count       = dispatchStats[b.id] || 0;
  const isLastPicked = b.id === lastPicked;
  const lastSeen    = new Date(b.last_seen).toLocaleTimeString();
  return ` + "`" + `
  <div class="backend${isLastPicked ? ' last-picked' : ''}">
    ${count > 0 ? ` + "`" + `<span class="dispatch-count">${count} req</span>` + "`" + ` : ''}
    <div class="backend-id"><span class="pulse ${pulseClass}"></span>${b.id}</div>
    <div class="backend-meta">
      addr &nbsp;&nbsp;${b.addr}<br>
      dc &nbsp;&nbsp;&nbsp;&nbsp;${b.dc}<br>
      weight &nbsp;${b.weight}<br>
      seen &nbsp;&nbsp;${lastSeen}
    </div>
    <div><span class="badge ${badgeClass}">${b.health}</span></div>
    <div class="actions">
      ${b.health !== 'healthy'   ? ` + "`" + `<button class="btn btn-green"  onclick="setHealth('${cluster}','${b.id}','healthy')">healthy</button>` + "`" + ` : ''}
      ${b.health !== 'unhealthy' ? ` + "`" + `<button class="btn btn-red"    onclick="setHealth('${cluster}','${b.id}','unhealthy')">unhealthy</button>` + "`" + ` : ''}
      ${b.health !== 'draining'  ? ` + "`" + `<button class="btn btn-yellow" onclick="setHealth('${cluster}','${b.id}','draining')">drain</button>` + "`" + ` : ''}
      <button class="btn btn-gray" onclick="removeBackend('${cluster}','${b.id}')">remove</button>
    </div>
  </div>` + "`" + `;
}

async function setHealth(cluster, id, health) {
  try {
    const res = await fetch(` + "`" + `/clusters/${cluster}/backends/${id}/health` + "`" + `, {
      method: 'PATCH', headers: {'Content-Type': 'application/json'}, body: JSON.stringify({health}),
    });
    if (!res.ok) throw new Error(await res.text());
    load();
  } catch(e) { showError(e.message); }
}

async function removeBackend(cluster, id) {
  try {
    await fetch(` + "`" + `/clusters/${cluster}/backends/${id}` + "`" + `, {method: 'DELETE'});
    load();
  } catch(e) { showError(e.message); }
}

async function addBackend() {
  const cluster = document.getElementById('f-cluster').value.trim();
  const id      = document.getElementById('f-id').value.trim();
  const addr    = document.getElementById('f-addr').value.trim();
  const dc      = document.getElementById('f-dc').value.trim();
  const weight  = parseInt(document.getElementById('f-weight').value) || 1;
  if (!cluster || !id || !addr) { showError('cluster, id and addr are required'); return; }
  try {
    const res = await fetch(` + "`" + `/clusters/${cluster}/backends` + "`" + `, {
      method: 'POST', headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({id, addr, dc, weight}),
    });
    if (!res.ok) throw new Error((await res.json()).error);
    document.getElementById('f-id').value = '';
    document.getElementById('f-addr').value = '';
    load();
  } catch(e) { showError(e.message); }
}

function showError(msg) {
  const el = document.getElementById('error-bar');
  el.textContent = msg;
  el.style.display = 'block';
  setTimeout(() => el.style.display = 'none', 4000);
}

load();
setInterval(load, 2000);
</script>
</body>
</html>`
