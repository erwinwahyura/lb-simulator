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
  .backend-meta span[data-tip] { border-bottom: 1px dotted #484f58; cursor: help; }
  .dispatch-count { position: absolute; top: 10px; right: 10px; font-size: 11px; color: #58a6ff; background: #0d2040; border: 1px solid #1f4070; border-radius: 10px; padding: 1px 7px; cursor: help; }
  .dist-bar-wrap { margin: 6px 0 2px; height: 6px; background: #21262d; border-radius: 3px; overflow: hidden; }
  .dist-bar-fill { height: 100%; border-radius: 3px; transition: width .4s ease; background: #1f6feb; }
  .dist-pct { font-size: 10px; color: #58a6ff; margin-bottom: 8px; }
  .badge { display: inline-block; padding: 2px 7px; border-radius: 10px; font-size: 10px; font-weight: 600; margin-bottom: 8px; cursor: help; }
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
  .field label { font-size: 11px; color: #8b949e; cursor: help; }
  .field label[data-tip] { border-bottom: 1px dotted #484f58; }
  .field input { background: #0d1117; border: 1px solid #30363d; color: #e6edf3; padding: 5px 9px; border-radius: 4px; font-family: inherit; font-size: 12px; width: 140px; }
  .field input.short { width: 80px; }
  .field input:focus { outline: none; border-color: #58a6ff; }
  .algo-group { display: flex; gap: 5px; }
  .algo-btn { padding: 5px 10px; font-size: 11px; font-family: inherit; border-radius: 4px; border: 1px solid #30363d; cursor: pointer; background: #21262d; color: #8b949e; position: relative; }
  .algo-btn.active { background: #0d2040; color: #58a6ff; border-color: #1f4070; }
  .algo-btn:hover { border-color: #58a6ff44; }
  .result-box { margin-top: 12px; padding: 10px 12px; background: #0d1117; border: 1px solid #30363d; border-radius: 6px; font-size: 12px; min-height: 38px; color: #8b949e; }
  /* dispatcher */
  .disp-stats { display: flex; gap: 24px; margin-top: 16px; }
  .disp-stat { display: flex; flex-direction: column; gap: 2px; cursor: help; }
  .disp-stat-val { font-size: 22px; font-weight: 700; color: #f0f6fc; }
  .disp-stat-lbl { font-size: 10px; color: #8b949e; text-transform: uppercase; letter-spacing: .06em; border-bottom: 1px dotted #484f58; display: inline-block; }
  .disp-bars { display: flex; flex-direction: column; gap: 8px; margin-top: 14px; }
  .disp-row { display: grid; grid-template-columns: 130px 1fr 60px; gap: 8px; align-items: center; font-size: 11px; }
  .disp-row-id { color: #e6edf3; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .disp-bar-bg { background: #21262d; border-radius: 3px; height: 8px; overflow: hidden; cursor: help; }
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

  /* ── l7 routes ── */
  .routes-table { width: 100%; border-collapse: collapse; font-size: 12px; margin-bottom: 14px; }
  .routes-table th { text-align: left; padding: 6px 10px; color: #8b949e; font-weight: 500; border-bottom: 1px solid #30363d; font-size: 11px; }
  .routes-table td { padding: 6px 10px; border-bottom: 1px solid #21262d; color: #e6edf3; }
  .routes-table tr:last-child td { border-bottom: none; }
  .routes-table .priority { color: #58a6ff; }
  .routes-table .cluster-tag { background: #0d2040; color: #58a6ff; border: 1px solid #1f4070; border-radius: 4px; padding: 1px 6px; font-size: 10px; }
  .routes-table .cluster-tag.dead { background: #3a1a1a; color: #f85149; border-color: #da3633; }

  /* ── pipeline ── */
  .trace-box { background: #0d1117; border: 1px solid #30363d; border-radius: 6px; padding: 14px; margin-top: 12px; font-size: 11px; line-height: 1.7; }
  .trace-box .ok    { color: #3fb950; }
  .trace-box .err   { color: #f85149; }
  .trace-box .warn  { color: #d29922; }
  .trace-box .dim   { color: #484f58; }
  .trace-box .label { color: #8b949e; display: inline-block; width: 140px; }
  .trace-log { margin-top: 14px; }
  .trace-log-title { font-size: 11px; color: #8b949e; margin-bottom: 6px; }
  .trace-row { display: grid; grid-template-columns: 70px 50px 1fr 90px 90px 60px 60px; gap: 6px; align-items: center; font-size: 11px; padding: 4px 0; border-bottom: 1px solid #21262d; }
  .trace-row.conflict { background: #1a1000; }
  .trace-row.error    { background: #1a0808; }
  .trace-row .method  { color: #58a6ff; font-weight: 600; }
  .trace-row .path    { color: #e6edf3; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .trace-row .tag     { padding: 1px 5px; border-radius: 3px; font-size: 10px; text-align: center; }
  .tag-ok      { background: #1a3a2a; color: #3fb950; }
  .tag-conflict{ background: #3a2e1a; color: #d29922; }
  .tag-error   { background: #3a1a1a; color: #f85149; }
  .trace-log-header { display: grid; grid-template-columns: 70px 50px 1fr 90px 90px 60px 60px; gap: 6px; font-size: 10px; color: #484f58; padding: 4px 0; border-bottom: 1px solid #30363d; text-transform: uppercase; }

  /* ── geo routing ── */
  .region-grid { display: flex; flex-wrap: wrap; gap: 8px; margin-bottom: 14px; }
  .region-btn { padding: 8px 14px; font-size: 12px; font-family: inherit; border-radius: 6px; border: 1px solid #30363d; cursor: pointer; background: #21262d; color: #e6edf3; text-align: left; }
  .region-btn:hover { border-color: #58a6ff44; }
  .region-btn .flag { font-size: 16px; margin-right: 6px; }
  .geo-result { background: #0d1117; border: 1px solid #30363d; border-radius: 6px; padding: 12px 14px; font-size: 12px; min-height: 48px; color: #8b949e; }
  .geo-scores { display: flex; gap: 10px; margin-top: 10px; flex-wrap: wrap; }
  .geo-score { padding: 4px 10px; border-radius: 4px; font-size: 11px; border: 1px solid #30363d; }
  .geo-score.chosen  { border-color: #238636; background: #1a3a2a; color: #3fb950; }
  .geo-score.down    { border-color: #484f58; background: #161b22; color: #484f58; text-decoration: line-through; }
  .geo-score.other   { background: #161b22; color: #8b949e; }

  /* ── capacity test ── */
  .cap-result { margin-top: 14px; font-size: 12px; }
  .cap-big { font-size: 28px; font-weight: 700; color: #f0f6fc; }
  .cap-label { font-size: 10px; color: #8b949e; text-transform: uppercase; letter-spacing: .06em; }
  .cap-grid { display: flex; gap: 28px; margin-top: 10px; }
  .cap-note { font-size: 11px; color: #484f58; margin-top: 8px; line-height: 1.5; }
  .progress-bar { height: 6px; background: #21262d; border-radius: 3px; margin-top: 10px; overflow: hidden; }
  .progress-fill { height: 100%; background: #1f6feb; border-radius: 3px; transition: width .1s; }

  /* ── failover lab ── */
  .dc-grid { display: flex; flex-wrap: wrap; gap: 12px; margin-bottom: 14px; }
  .dc-card { background: #0d1117; border: 1px solid #30363d; border-radius: 6px; padding: 12px 16px; min-width: 200px; }
  .dc-card.dc-impaired { border-color: #da363388; }
  .dc-card.dc-draining { border-color: #9e6a0388; }
  .dc-name  { font-size: 13px; font-weight: 600; color: #f0f6fc; margin-bottom: 6px; }
  .dc-count { font-size: 11px; color: #8b949e; margin-bottom: 10px; }
  .dc-count .ok { color: #3fb950; }
  .dc-count .bad { color: #f85149; }
  .dc-count .drain { color: #d29922; }
  .dc-actions { display: flex; gap: 6px; }
  .event-log { margin-top: 14px; }
  .event-log-title { font-size: 11px; color: #8b949e; margin-bottom: 6px; }
  .event-list { list-style: none; font-size: 11px; color: #8b949e; max-height: 100px; overflow-y: auto; display: flex; flex-direction: column; gap: 3px; }
  .event-list li { padding: 3px 0; border-bottom: 1px solid #21262d; }
  .event-list .ev-time { color: #484f58; margin-right: 8px; }
  .event-list .ev-fail { color: #f85149; }
  .event-list .ev-drain { color: #d29922; }
  .event-list .ev-ok   { color: #3fb950; }

  /* ── tooltip ── */
  #tip {
    position: fixed;
    background: #1c2128;
    border: 1px solid #444c56;
    border-radius: 6px;
    padding: 9px 13px;
    font-size: 12px;
    line-height: 1.55;
    color: #e6edf3;
    max-width: 280px;
    pointer-events: none;
    opacity: 0;
    transition: opacity .12s;
    z-index: 9999;
    box-shadow: 0 6px 20px rgba(0,0,0,.5);
    white-space: pre-wrap;
  }
  #tip.show { opacity: 1; }
  #tip .tip-title { color: #58a6ff; font-weight: 600; margin-bottom: 4px; font-size: 11px; }
</style>
</head>
<body>
<div id="tip"></div>

<h1>lbsim — Control Plane</h1>
<p class="subtitle">L1–L7 · registry · health · routing · dispatcher · failover · geo · L7 pipeline · hover any element for explanation</p>

<div id="error-bar" class="error-bar"></div>

<!-- ── Layer 4: Dispatcher ───────────────────────── -->
<div class="panel">
  <h2><span id="disp-dot" class="stopped-dot"></span>Live Dispatcher — Layer 4</h2>
  <div class="form-row">
    <div class="field">
      <label data-tip="Which cluster to send traffic to. A cluster is a pool of interchangeable backends (replicas of one service).">cluster</label>
      <input id="d-cluster" value="api" class="short">
    </div>
    <div class="field">
      <label data-tip="The algorithm the router uses to pick a backend for each request. Each algorithm has different trade-offs — hover the buttons to learn.">algorithm</label>
      <div class="algo-group" id="d-algo-group">
        <button class="algo-btn active" onclick="setDispAlgo('round-robin',this)"
          data-tip="Round-Robin&#10;&#10;Cycles through backends in order — backend A, then B, then C, then A again. Every backend gets exactly one request before any gets a second.&#10;&#10;✓ Simple and fair&#10;✓ Good when all backends have equal capacity&#10;✗ Ignores weight differences">round-robin</button>
        <button class="algo-btn" onclick="setDispAlgo('weighted',this)"
          data-tip="Weighted Random&#10;&#10;Backends receive traffic proportional to their weight. Imagine a number line 0→total_weight; each request picks a random point and lands on whichever backend owns that segment.&#10;&#10;tokyo-api-01 (w:5) gets ~38% of traffic&#10;tokyo-api-02 (w:5) gets ~38%&#10;osaka-api-01  (w:3) gets ~23%&#10;&#10;✓ Matches capacity differences&#10;✓ Good for heterogeneous hardware">weighted</button>
        <button class="algo-btn" onclick="setDispAlgo('consistent-hash',this)"
          data-tip="Consistent Hashing&#10;&#10;Maps a key (e.g. user ID) to a position on a virtual ring. The backend owning the nearest ring slot serves that key — always the same one.&#10;&#10;Example: user-42 → always hits tokyo-api-01&#10;         user-99 → always hits osaka-api-01&#10;&#10;✓ Session affinity (same user, same server)&#10;✓ Cache locality (hot data stays on one backend)&#10;✗ Uneven if few backends; 150 virtual nodes mitigate this">consistent-hash</button>
      </div>
    </div>
    <div class="field">
      <label data-tip="Requests per second the dispatcher fires through the router. Each tick picks one healthy backend via the selected algorithm.">req/s</label>
      <input id="d-rps" value="5" type="number" min="1" max="50" class="short">
    </div>
    <div class="field"><label>&nbsp;</label>
      <button class="btn btn-blue" onclick="dispStart()"
        data-tip="Start firing requests at the configured rate. Watch the traffic bars grow and try marking a backend unhealthy to see failover in action.">▶ Start</button>
    </div>
    <div class="field"><label>&nbsp;</label>
      <button class="btn btn-stop" onclick="dispStop()"
        data-tip="Pause the dispatcher. Counters are preserved — click Start again to resume.">■ Stop</button>
    </div>
    <div class="field"><label>&nbsp;</label>
      <button class="btn btn-gray" onclick="dispReset()"
        data-tip="Stop the dispatcher and clear all traffic counters back to zero.">↺ Reset</button>
    </div>
  </div>

  <div class="disp-stats">
    <div class="disp-stat" data-tip="Total requests fired since last reset. Each request is routed to exactly one healthy backend.">
      <span class="disp-stat-val" id="d-total">0</span>
      <span class="disp-stat-lbl">total requests</span>
    </div>
    <div class="disp-stat" data-tip="Rolling average over the last 10 seconds. Stabilises after ~10s of running. Will drop if backends go unhealthy (skipped ticks).">
      <span class="disp-stat-val" id="d-rps-live">0.0</span>
      <span class="disp-stat-lbl">req/s (10s avg)</span>
    </div>
    <div class="disp-stat" data-tip="The algorithm currently in use by the running dispatcher.">
      <span class="disp-stat-val" id="d-algo-live">—</span>
      <span class="disp-stat-lbl">algorithm</span>
    </div>
  </div>

  <div id="disp-bars" class="disp-bars"></div>
</div>

<!-- ── Layer 7: L7 Route Rules ───────────────────── -->
<div class="panel">
  <h2>L7 Route Rules — Layer 7</h2>
  <p style="font-size:11px;color:#8b949e;margin-bottom:12px"
     data-tip="L7 routing reads the HTTP request (path, method, headers) before picking a cluster.&#10;&#10;Rules are checked in priority order (lowest first). First match wins.&#10;&#10;Without L7 rules, all requests go to one cluster regardless of path. With L7 rules, /api/* can go to the api cluster while /static/* goes to a CDN or static-file cluster.">
    Rules are matched in priority order. First match wins. L7 runs after L4 — see the Pipeline below to watch them interact.
  </p>
  <table class="routes-table" id="routes-table">
    <thead><tr>
      <th data-tip="Execution order — lower priority number = checked first.">priority</th>
      <th data-tip="HTTP method filter. * matches any method.">method</th>
      <th data-tip="Request path must start with this prefix to match.">path prefix</th>
      <th data-tip="The cluster requests are sent to when this rule matches.">→ cluster</th>
      <th></th>
    </tr></thead>
    <tbody id="routes-body"><tr><td colspan="5" style="color:#484f58;padding:12px 10px">loading…</td></tr></tbody>
  </table>
  <div class="form-row">
    <div class="field"><label data-tip="Lower number = higher priority. Use 10 for specific routes, 99 for catch-all.">priority</label><input id="rt-priority" value="10" class="short" type="number"></div>
    <div class="field"><label data-tip="HTTP method. Use * to match all methods.">method</label><input id="rt-method" value="*" class="short"></div>
    <div class="field"><label data-tip="Path prefix to match. Must start with /. Use / as a catch-all fallback.">path prefix</label><input id="rt-path" placeholder="/api/" style="width:120px"></div>
    <div class="field"><label data-tip="Which cluster to route matching requests to. Must exist in the registry.">cluster</label><input id="rt-cluster" placeholder="api" class="short"></div>
    <div class="field"><label>&nbsp;</label><button class="btn btn-green" onclick="addRoute()" data-tip="Add this rule to the L7 routing table.">+ Add Rule</button></div>
  </div>
</div>

<!-- ── Layer 7: Request Pipeline ─────────────────── -->
<div class="panel">
  <h2>Request Pipeline — L4 + L7 + Conflict Detection</h2>
  <p style="font-size:11px;color:#8b949e;margin-bottom:12px"
     data-tip="A request enters the LB stack at L4 (TCP level) then L7 (HTTP level).&#10;&#10;L4 sees only client IP and port — it hashes the IP to pick a backend cluster.&#10;L7 sees the HTTP path — it matches against route rules to pick a cluster.&#10;&#10;Conflict: when L4 and L7 pick DIFFERENT clusters. L7 always wins, but if the L7 cluster has no backends, the request fails.&#10;&#10;Try: fire a request to /static/logo.png — L4 hashes your IP to 'api', L7 matches /static/ → 'static' (no backends). Conflict + error.">
    Fire a request and watch it travel through L4 → L7 → backend. Try <strong style="color:#d29922">/static/logo.png</strong> to trigger an OSI conflict.
  </p>
  <div class="form-row">
    <div class="field">
      <label data-tip="HTTP method sent with the request.">method</label>
      <select id="pipe-method" style="background:#0d1117;border:1px solid #30363d;color:#e6edf3;padding:5px 9px;border-radius:4px;font-family:inherit;font-size:12px">
        <option>GET</option><option>POST</option><option>PUT</option><option>DELETE</option>
      </select>
    </div>
    <div class="field"><label data-tip="The request path. L7 matches this against route rules. Try /api/users, /static/logo.png, /unknown.">path</label><input id="pipe-path" value="/api/users" style="width:180px"></div>
    <div class="field"><label data-tip="Simulated client IP. L4 hashes this to pick a cluster. Change it to see different L4 decisions.">client IP</label><input id="pipe-ip" value="10.0.0.42" style="width:120px"></div>
    <div class="field"><label>&nbsp;</label><button class="btn btn-blue" onclick="processPipeline()" data-tip="Send this request through the full L4 + L7 pipeline and show the trace.">▶ Process</button></div>
    <div class="field"><label>&nbsp;</label><button class="btn btn-gray" onclick="pipelineScenario('conflict')" data-tip="Pre-fill a request that causes an L4/L7 conflict: L4 picks 'api', L7 matches /static/ → 'static' (no backends).">⚠ Conflict demo</button></div>
  </div>
  <div id="trace-box" class="trace-box" style="color:#484f58">fire a request to see the pipeline trace</div>
  <div class="trace-log" id="trace-log-wrap" style="display:none">
    <div class="trace-log-title">request history (last 20)</div>
    <div class="trace-log-header"><span>time</span><span>method</span><span>path</span><span>L4 cluster</span><span>L7 cluster</span><span>backend</span><span>status</span></div>
    <div id="trace-log"></div>
  </div>
</div>

<!-- ── Layer 6: Geo Routing ──────────────────────── -->
<div class="panel">
  <h2>Geo Routing — Layer 6</h2>
  <p style="font-size:11px;color:#8b949e;margin-bottom:12px"
     data-tip="GeoDNS / latency-based DC selection&#10;&#10;Before picking a backend, the geo router first picks WHICH datacenter to route to — based on where the user is. Closest DC with healthy backends wins.&#10;&#10;If the preferred DC goes down, it automatically falls back to the next-closest one. This is the 'global traffic management' layer above round-robin.">
    Click a region to route a request. The geo router picks the lowest-latency healthy DC, then round-robins within it.
  </p>
  <div class="region-grid">
    <button class="region-btn" onclick="geoRoute('japan')"
      data-tip="Japan → prefers tokyo [8ms] over osaka [25ms]&#10;&#10;Try: fail the tokyo DC, then click Japan again. The geo router falls back to osaka automatically.">
      <span class="flag">🗾</span>Japan
    </button>
    <button class="region-btn" onclick="geoRoute('korea')"
      data-tip="Korea → prefers osaka [12ms] over tokyo [30ms]&#10;&#10;Korea is closer to osaka geographically, so osaka wins even though both are in Japan.">
      <span class="flag">🇰🇷</span>Korea
    </button>
    <button class="region-btn" onclick="geoRoute('us')"
      data-tip="US → both DCs are far (tokyo 105ms, osaka 110ms)&#10;&#10;US traffic goes to the slightly closer DC. In a real system you'd add a us-west DC. Try adding one via Register Backend above.">
      <span class="flag">🇺🇸</span>US
    </button>
    <button class="region-btn" onclick="geoRoute('europe')"
      data-tip="Europe → prefers osaka [190ms] over tokyo [195ms]&#10;&#10;Both are very far. 5ms difference. In production you'd add a europe DC. Without one, geo routing still works — it just picks the least-bad option.">
      <span class="flag">🇪🇺</span>Europe
    </button>
    <button class="region-btn" onclick="geoRoute('australia')"
      data-tip="Australia → prefers osaka [75ms] over tokyo [80ms]&#10;&#10;Closest of all non-Japan regions to the Japan DCs.">
      <span class="flag">🇦🇺</span>Australia
    </button>
  </div>
  <div class="geo-result" id="geo-result">click a region to see latency-based DC selection</div>
</div>

<!-- ── Capacity Test ──────────────────────────────── -->
<div class="panel">
  <h2>Server Capacity</h2>
  <p style="font-size:11px;color:#8b949e;margin-bottom:12px"
     data-tip="Two numbers matter:&#10;&#10;Internal RPS — how fast the routing logic runs in pure Go (no HTTP). Measures the mutex + map lookup speed.&#10;&#10;HTTP RPS — how many real HTTP requests the server handles per second from the browser. Includes network stack, JSON encode/decode, and Go's HTTP mux overhead.&#10;&#10;The gap between the two is the HTTP overhead cost.">
    Two measurements: pure routing speed (internal) vs real HTTP throughput.
  </p>
  <div class="form-row">
    <div class="field">
      <label data-tip="How many route picks to run for the internal benchmark. Higher = more accurate, takes longer.">internal n</label>
      <input id="cap-n" value="50000" class="short" type="number">
    </div>
    <div class="field">
      <label data-tip="Concurrent browser fetch() calls for the HTTP test. Limited by browser (~6 per host by default, but fetch uses keep-alive so higher concurrency is fine).">http concurrency</label>
      <input id="cap-conc" value="20" class="short" type="number">
    </div>
    <div class="field">
      <label data-tip="How long to run the HTTP benchmark in seconds.">http duration (s)</label>
      <input id="cap-dur" value="3" class="short" type="number">
    </div>
    <div class="field"><label>&nbsp;</label>
      <button class="btn btn-blue" onclick="runCapTest()"
        data-tip="Run both benchmarks and show the results. The internal benchmark completes in milliseconds. The HTTP benchmark runs for the configured duration.">▶ Run Test</button>
    </div>
  </div>
  <div class="progress-bar"><div class="progress-fill" id="cap-progress" style="width:0%"></div></div>
  <div class="cap-result" id="cap-result" style="color:#484f58">run the test to see capacity numbers</div>
</div>

<!-- ── Layer 5: Failover Lab ────────────────────── -->
<div class="panel">
  <h2>Failover Lab — Layer 5</h2>
  <div id="dc-grid" class="dc-grid"></div>
  <div class="event-log">
    <div class="event-log-title">event log</div>
    <ul id="event-list" class="event-list"><li style="color:#484f58">no events yet — try failing a DC</li></ul>
  </div>
</div>

<!-- ── Layer 3: Manual routing ──────────────────── -->
<div class="panel">
  <h2>Manual Route — Layer 3</h2>
  <div class="form-row">
    <div class="field">
      <label data-tip="The cluster to route into. All healthy backends in this cluster are candidates.">cluster</label>
      <input id="r-cluster" value="api" class="short">
    </div>
    <div class="field">
      <label data-tip="Algorithm used to pick one backend from the healthy pool.">algorithm</label>
      <div class="algo-group">
        <button class="algo-btn active" onclick="setRouteAlgo('round-robin',this)"
          data-tip="Round-Robin&#10;&#10;Cycles through backends in order — backend A, then B, then C, then A again. Every backend gets exactly one request before any gets a second.&#10;&#10;✓ Simple and fair&#10;✓ Good when all backends have equal capacity&#10;✗ Ignores weight differences">round-robin</button>
        <button class="algo-btn" onclick="setRouteAlgo('weighted',this)"
          data-tip="Weighted Random&#10;&#10;Backends receive traffic proportional to their weight. Imagine a number line 0→total_weight; each request picks a random point and lands on whichever backend owns that segment.&#10;&#10;tokyo-api-01 (w:5) gets ~38% of traffic&#10;tokyo-api-02 (w:5) gets ~38%&#10;osaka-api-01  (w:3) gets ~23%&#10;&#10;✓ Matches capacity differences&#10;✓ Good for heterogeneous hardware">weighted</button>
        <button class="algo-btn" onclick="setRouteAlgo('consistent-hash',this)"
          data-tip="Consistent Hashing&#10;&#10;Maps a key (e.g. user ID) to a position on a virtual ring. The backend owning the nearest ring slot serves that key — always the same one.&#10;&#10;Example: user-42 → always hits tokyo-api-01&#10;         user-99 → always hits osaka-api-01&#10;&#10;✓ Session affinity (same user, same server)&#10;✓ Cache locality (hot data stays on one backend)&#10;✗ Uneven if few backends; 150 virtual nodes mitigate this">consistent-hash</button>
      </div>
    </div>
    <div class="field">
      <label data-tip="Only used by consistent-hash. The same key always maps to the same backend. Try 'user-42' repeatedly — it always lands on the same backend.">key</label>
      <input id="r-key" placeholder="user-123" class="short">
    </div>
    <div class="field"><label>&nbsp;</label>
      <button class="btn btn-blue" onclick="manualDispatch(1)"
        data-tip="Send one request through the router and show which backend was selected.">Dispatch</button>
    </div>
    <div class="field"><label>&nbsp;</label>
      <button class="btn btn-gray" onclick="manualDispatch(20)"
        data-tip="Send 20 requests in a row. Good for seeing weighted distribution or verifying consistent-hash stickiness.">× 20</button>
    </div>
  </div>
  <div class="result-box" id="result-box">pick an algorithm and click Dispatch</div>
</div>

<!-- ── Layer 1: Register backend ────────────────── -->
<div class="panel">
  <h2>Register Backend — Layer 1</h2>
  <div class="form-row">
    <div class="field">
      <label data-tip="Which cluster this backend belongs to. A cluster groups backends that serve the same service (e.g. all 'api' replicas).">cluster</label>
      <input id="f-cluster" value="api">
    </div>
    <div class="field">
      <label data-tip="Stable identity for this backend. Used as the key in the registry and in consistent-hash ring placement.">id</label>
      <input id="f-id" placeholder="singapore-api-01">
    </div>
    <div class="field">
      <label data-tip="IP:port where this backend lives. The health checker would probe this address; the data plane would forward packets here.">addr</label>
      <input id="f-addr" placeholder="10.0.3.1:8080">
    </div>
    <div class="field">
      <label data-tip="Datacenter this backend is in. Used in Layer 5 (failover) to group backends by location.">dc</label>
      <input id="f-dc" placeholder="singapore" class="short">
    </div>
    <div class="field">
      <label data-tip="Relative traffic weight. A backend with weight 5 receives ~5× more traffic than one with weight 1 under the weighted algorithm. Ignored by round-robin.">weight</label>
      <input id="f-weight" placeholder="3" class="short" type="number">
    </div>
    <div class="field"><label>&nbsp;</label>
      <button class="btn btn-green" onclick="addBackend()"
        data-tip="Register this backend into the cluster. It starts healthy immediately.">+ Register</button>
    </div>
  </div>
</div>

<!-- ── Cluster state ─────────────────────────────── -->
<div id="clusters"></div>
<p class="ticker" id="ticker">refreshing...</p>

<script>
let routeAlgo  = 'round-robin';
let dispAlgo   = 'round-robin';
let dispStats  = {};
let routeStats = {};
let lastPicked = null;
let lastClusters = [];

// ── tooltip ─────────────────────────────────────────
const tip = document.getElementById('tip');
let tipVisible = false;

document.addEventListener('mousemove', e => {
  if (tipVisible) {
    let x = e.clientX + 16, y = e.clientY - 10;
    if (x + 290 > window.innerWidth)  x = e.clientX - 300;
    if (y + tip.offsetHeight > window.innerHeight) y = e.clientY - tip.offsetHeight - 10;
    tip.style.left = x + 'px';
    tip.style.top  = y + 'px';
  }
});

document.addEventListener('mouseover', e => {
  const el = e.target.closest('[data-tip]');
  if (!el) return;
  const raw = el.getAttribute('data-tip');
  const lines = raw.split('&#10;');
  const first = lines[0];
  const rest  = lines.slice(1).join('\n').trimStart();
  tip.innerHTML = rest
    ? ` + "`" + `<div class="tip-title">${first}</div><div>${rest}</div>` + "`" + `
    : ` + "`" + `<div>${first}</div>` + "`" + `;
  tip.classList.add('show');
  tipVisible = true;
});

document.addEventListener('mouseout', e => {
  const el = e.target.closest('[data-tip]');
  if (!el) return;
  tip.classList.remove('show');
  tipVisible = false;
});

// ── dc colour ────────────────────────────────────────
function dcClass(dc) {
  return dc === 'tokyo' ? 'disp-bar-dc-tokyo' : dc === 'osaka' ? 'disp-bar-dc-osaka' : 'disp-bar-dc-other';
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

async function dispStop()  { await fetch('/dispatcher/stop',  {method:'POST'}); }
async function dispReset() {
  await fetch('/dispatcher/reset', {method:'POST'});
  dispStats = {};
  renderDispBars({backends:[], total:0, rps:0, algo:'—', running:false});
}

async function loadDispStats() {
  try {
    const res = await fetch('/dispatcher/stats');
    dispStats  = await res.json();
    renderDispBars(dispStats);
  } catch(_) {}
}

function renderDispBars(s) {
  document.getElementById('d-total').textContent     = (s.total || 0).toLocaleString();
  document.getElementById('d-rps-live').textContent  = (s.rps   || 0).toFixed(1);
  document.getElementById('d-algo-live').textContent = s.algo || '—';
  document.getElementById('disp-dot').className      = s.running ? 'running-dot' : 'stopped-dot';

  const backends = s.backends || [];
  const total    = s.total    || 0;
  const wrap     = document.getElementById('disp-bars');
  if (!backends.length) { wrap.innerHTML = ''; return; }

  backends.sort((a,b) => b.total - a.total);
  wrap.innerHTML = backends.map(b => {
    const pct = total > 0 ? ((b.total / total)*100).toFixed(1) : '0.0';
    const cls = dcClass(b.dc);
    return ` + "`" + `
    <div class="disp-row">
      <div class="disp-row-id" title="${b.id}">${b.id}</div>
      <div class="disp-bar-bg"
           data-tip="Traffic share for ${b.id} (dc: ${b.dc}). ${pct}% of all dispatched requests landed here.">
        <div class="disp-bar-fg ${cls}" style="width:${pct}%"></div>
      </div>
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
  const key     = document.getElementById('r-key').value.trim()     || 'default';
  let last;
  for (let i = 0; i < times; i++) {
    try {
      const res = await fetch('/route', {
        method:'POST', headers:{'Content-Type':'application/json'},
        body: JSON.stringify({cluster, algo: routeAlgo, key}),
      });
      if (!res.ok) { throw new Error((await res.json()).error); }
      last = await res.json();
    } catch(e) { showError(e.message); return; }
  }
  if (last) {
    lastPicked = last.backend.id;
    const b   = last.backend;
    const sfx = times > 1 ? ` + "`" + ` (after ${times} dispatches)` + "`" + ` : '';
    document.getElementById('result-box').innerHTML =
      ` + "`" + `→ <strong style="color:#3fb950">${b.id}</strong>&nbsp; dc:${b.dc} weight:${b.weight} algo:${last.algo}${sfx}` + "`" + `;
    await loadRouteStats();
    render(lastClusters);
  }
}

async function loadRouteStats() {
  try { routeStats = await (await fetch('/route/stats')).json(); } catch(_) {}
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
    document.getElementById('f-id').value = document.getElementById('f-addr').value = '';
    load();
  } catch(e) { showError(e.message); }
}

// ── cluster rendering ────────────────────────────────
async function load() {
  try {
    const [cRes] = await Promise.all([fetch('/clusters'), loadRouteStats(), loadDispStats(), loadRoutes(), loadTraces()]);
    if (!cRes.ok) throw new Error(await cRes.text());
    lastClusters = await cRes.json();
    knownClusters = lastClusters.map(c => c.name);
    render(lastClusters);
    document.getElementById('error-bar').style.display = 'none';
    document.getElementById('ticker').textContent = 'last updated: ' + new Date().toLocaleTimeString();
  } catch(e) { showError('fetch failed: ' + e.message); }
}

function render(clusters) {
  renderDCGrid(clusters);
  const el = document.getElementById('clusters');
  if (!clusters || !clusters.length) { el.innerHTML = '<p class="empty">no clusters</p>'; return; }
  el.innerHTML = clusters.map(c => {
    const backends = Object.values(c.backends || {});
    const healthy  = backends.filter(b => b.health === 'healthy').length;
    return ` + "`" + `
    <div class="cluster">
      <div class="cluster-header">
        <span class="cluster-name">${c.name}</span>
        <span class="cluster-count"
          data-tip="Healthy backends are eligible for routing. Unhealthy/draining ones are skipped by the router.">${healthy}/${backends.length} healthy</span>
      </div>
      <div class="backends">
        ${backends.length ? backends.map(b => backendCard(c.name, b)).join('') : '<p class="empty">no backends</p>'}
      </div>
    </div>` + "`" + `;
  }).join('');
}

function backendCard(cluster, b) {
  const pulseClass = {healthy:'pulse-green', unhealthy:'pulse-red', draining:'pulse-yellow'}[b.health] || 'pulse-red';
  const badgeTip   = {
    healthy:   'Healthy — this backend is eligible to receive new requests. The router will select it.',
    unhealthy: 'Unhealthy — failed health checks. The router skips this backend. It can recover and come back automatically.',
    draining:  'Draining — graceful removal. No new requests are sent here, but any in-flight requests finish. Safer than hard-marking unhealthy.',
  }[b.health] || '';
  const badgeCls = {healthy:'badge-healthy', unhealthy:'badge-unhealthy', draining:'badge-draining'}[b.health] || '';
  const rCount   = routeStats[b.id] || 0;
  const dStat    = (dispStats.backends||[]).find(x => x.id === b.id);
  const dTotal   = dStat ? dStat.total : 0;
  const gTotal   = dispStats.total || 0;
  const pct      = gTotal > 0 ? ((dTotal / gTotal)*100).toFixed(1) : null;
  const lastSeen = new Date(b.last_seen).toLocaleTimeString();
  return ` + "`" + `
  <div class="backend${b.id === lastPicked ? ' last-picked' : ''}">
    ${rCount > 0 ? ` + "`" + `<span class="dispatch-count" data-tip="Manual dispatch count — total requests routed here via the 'Route a request' panel.">${rCount} req</span>` + "`" + ` : ''}
    <div class="backend-id"><span class="pulse ${pulseClass}"></span>${b.id}</div>
    <div class="backend-meta">
      addr &nbsp;${b.addr}<br>
      dc &nbsp;&nbsp;&nbsp;${b.dc}<br>
      <span data-tip="Relative traffic weight. Used by the weighted algorithm. Higher = more traffic share.">weight ${b.weight}</span><br>
      <span data-tip="Timestamp of the last health check probe. Updates every 3 seconds while the checker runs.">seen &nbsp;${lastSeen}</span>
    </div>
    ${pct !== null ? ` + "`" + `
      <div class="dist-bar-wrap" data-tip="Share of live dispatcher traffic. ${pct}% of all requests landed here.">
        <div class="dist-bar-fill" style="width:${pct}%"></div>
      </div>
      <div class="dist-pct">${dTotal.toLocaleString()} dispatched — ${pct}% of traffic</div>
    ` + "`" + ` : ''}
    <div><span class="badge ${badgeCls}" data-tip="${badgeTip}">${b.health}</span></div>
    <div class="actions">
      ${b.health !== 'healthy'   ? ` + "`" + `<button class="btn btn-green"  onclick="setHealth('${cluster}','${b.id}','healthy')"   data-tip="Mark healthy — router will start sending traffic here again immediately.">healthy</button>` + "`" + ` : ''}
      ${b.health !== 'unhealthy' ? ` + "`" + `<button class="btn btn-red"    onclick="setHealth('${cluster}','${b.id}','unhealthy')" data-tip="Mark unhealthy — router skips this backend on the next request. Simulates a failed health check.">unhealthy</button>` + "`" + ` : ''}
      ${b.health !== 'draining'  ? ` + "`" + `<button class="btn btn-yellow" onclick="setHealth('${cluster}','${b.id}','draining')"  data-tip="Start draining — no new requests, in-flight ones finish. The graceful removal path used during deploys.">drain</button>` + "`" + ` : ''}
      <button class="btn btn-gray" onclick="removeBackend('${cluster}','${b.id}')" data-tip="Hard-remove this backend from the registry. Immediate — no draining. Use 'drain' first for a safe removal.">remove</button>
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

// ── L7 routes ────────────────────────────────────────
let knownClusters = [];

async function loadRoutes() {
  try {
    const res = await fetch('/routes');
    const rules = await res.json();
    renderRoutes(rules);
  } catch(_) {}
}

function renderRoutes(rules) {
  const tbody = document.getElementById('routes-body');
  if (!rules || !rules.length) {
    tbody.innerHTML = '<tr><td colspan="5" style="color:#484f58;padding:10px">no rules — add one below</td></tr>';
    return;
  }
  tbody.innerHTML = rules.map(r => {
    const hasBackends = knownClusters.includes(r.cluster);
    const clsCls = hasBackends ? 'cluster-tag' : 'cluster-tag dead';
    const clsTip = hasBackends ? r.cluster : r.cluster + ' (no backends — requests will fail)';
    return ` + "`" + `<tr>
      <td class="priority" data-tip="Priority ${r.priority} — checked ${r.priority < 50 ? 'early' : 'late'} in the rule list.">${r.priority}</td>
      <td>${r.method}</td>
      <td style="color:#e6edf3"><strong>${r.path_prefix}</strong></td>
      <td><span class="${clsCls}" data-tip="${clsTip}">${r.cluster}</span></td>
      <td><button class="btn btn-gray" onclick="deleteRoute('${r.id}')" style="padding:2px 7px;font-size:10px"
          data-tip="Remove rule ${r.id}. Requests that previously matched this rule will fall through to the next matching rule.">✕</button></td>
    </tr>` + "`" + `;
  }).join('');
}

async function addRoute() {
  const priority   = parseInt(document.getElementById('rt-priority').value) || 10;
  const method     = document.getElementById('rt-method').value.trim()  || '*';
  const pathPrefix = document.getElementById('rt-path').value.trim();
  const cluster    = document.getElementById('rt-cluster').value.trim();
  if (!pathPrefix || !cluster) { showError('path_prefix and cluster are required'); return; }
  try {
    const res = await fetch('/routes', {
      method: 'POST', headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({method, path_prefix: pathPrefix, cluster, priority}),
    });
    if (!res.ok) throw new Error((await res.json()).error);
    document.getElementById('rt-path').value = '';
    loadRoutes();
  } catch(e) { showError(e.message); }
}

async function deleteRoute(id) {
  await fetch(` + "`" + `/routes/${id}` + "`" + `, {method: 'DELETE'});
  loadRoutes();
}

// ── pipeline ──────────────────────────────────────────
function pipelineScenario(name) {
  if (name === 'conflict') {
    document.getElementById('pipe-method').value = 'GET';
    document.getElementById('pipe-path').value   = '/static/logo.png';
    document.getElementById('pipe-ip').value     = '10.0.0.42';
  }
}

async function processPipeline() {
  const method   = document.getElementById('pipe-method').value;
  const path     = document.getElementById('pipe-path').value.trim()   || '/';
  const clientIP = document.getElementById('pipe-ip').value.trim() || '10.0.0.1';
  try {
    const res = await fetch('/pipeline', {
      method: 'POST', headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({method, path, client_ip: clientIP}),
    });
    const t = await res.json();
    renderTrace(t);
    await loadTraces();
  } catch(e) { showError(e.message); }
}

function renderTrace(t) {
  const ok      = !t.error && !t.conflict;
  const backend = t.backend ? t.backend.id : '—';
  const latency = t.backend ? t.latency_ms + 'ms' : '—';

  const headerRows = Object.entries(t.headers || {}).map(([k,v]) =>
    ` + "`" + `<div><span class="label">${k}</span><span style="color:#e6edf3">${v}</span></div>` + "`" + `
  ).join('');

  const conflictNote = t.conflict
    ? ` + "`" + `<div class="warn">⚠ OSI CONFLICT — L4 picked cluster "<strong>${t.l4_cluster}</strong>" (by IP hash), L7 overrides to "<strong>${t.l7_cluster}</strong>" (by path rule). L7 wins — but if that cluster has no backends, the request fails.</div>` + "`" + `
    : '';
  const errorNote = t.error
    ? ` + "`" + `<div class="err">✗ ERROR: ${t.error}</div>` + "`" + `
    : ` + "`" + `<div class="ok">✓ routed to <strong>${backend}</strong> (${t.backend?.dc}, ${latency})</div>` + "`" + `;

  document.getElementById('trace-box').innerHTML = ` + "`" + `
    <div style="margin-bottom:10px">
      <span class="dim">${t.request_id}</span> &nbsp;
      <span style="color:#58a6ff;font-weight:600">${t.method}</span>
      <span style="color:#e6edf3"> ${t.path}</span>
      <span class="dim"> from ${t.client_ip}</span>
    </div>
    <div><span class="label">L4 decision</span>IP hash → cluster <strong>${t.l4_cluster||'none'}</strong> <span class="dim">(only sees IP:port)</span></div>
    <div><span class="label">L7 matched rule</span><span style="color:#d29922">${t.matched_rule}</span></div>
    <div><span class="label">L7 decision</span>path match → cluster <strong>${t.l7_cluster||'none'}</strong></div>
    ${conflictNote}
    ${errorNote}
    <details style="margin-top:10px">
      <summary style="cursor:pointer;color:#8b949e;font-size:11px">injected headers ▸</summary>
      <div style="margin-top:6px;padding-left:8px;border-left:2px solid #30363d">${headerRows}</div>
    </details>` + "`" + `;
}

async function loadTraces() {
  try {
    const res = await fetch('/pipeline/traces');
    const traces = await res.json();
    if (!traces || !traces.length) return;
    document.getElementById('trace-log-wrap').style.display = 'block';
    document.getElementById('trace-log').innerHTML = traces.slice(0,20).map(t => {
      const rowCls = t.error ? 'trace-row error' : t.conflict ? 'trace-row conflict' : 'trace-row';
      const status = t.error ? '<span class="tag tag-error">error</span>'
                   : t.conflict ? '<span class="tag tag-conflict">conflict</span>'
                   : '<span class="tag tag-ok">ok</span>';
      const be = t.backend ? t.backend.id : '—';
      return ` + "`" + `<div class="${rowCls}" data-tip="${t.request_id}: L4→${t.l4_cluster} L7→${t.l7_cluster} ${t.error||''}">
        <span class="dim">${t.timestamp}</span>
        <span class="method">${t.method}</span>
        <span class="path" title="${t.path}">${t.path}</span>
        <span class="dim">${t.l4_cluster||'—'}</span>
        <span class="dim">${t.l7_cluster||'—'}</span>
        <span class="dim">${be}</span>
        ${status}
      </div>` + "`" + `;
    }).join('');
  } catch(_) {}
}

// ── geo routing ──────────────────────────────────────
async function geoRoute(region) {
  try {
    const res = await fetch('/geo-route', {
      method: 'POST', headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({cluster: 'api', region}),
    });
    if (!res.ok) { const e = await res.json(); throw new Error(e.error); }
    const data = await res.json();
    const b = data.backend;
    const scores = data.dc_scores || [];
    const scoreHTML = scores.map(s => {
      const cls = !s.has_healthy ? 'geo-score down' : s.chosen ? 'geo-score chosen' : 'geo-score other';
      const suffix = !s.has_healthy ? ' ✗ down' : s.chosen ? ' ← chosen' : '';
      return ` + "`" + `<span class="${cls}" data-tip="${s.dc}: ${s.latency_ms}ms from ${region}${!s.has_healthy ? ' — no healthy backends, skipped' : ''}">${s.dc} ${s.latency_ms}ms${suffix}</span>` + "`" + `;
    }).join('');
    document.getElementById('geo-result').innerHTML =
      ` + "`" + `<strong style="color:#3fb950">${region}</strong> → <strong style="color:#f0f6fc">${b.id}</strong> (${b.dc}, weight:${b.weight})<div class="geo-scores">${scoreHTML}</div>` + "`" + `;
  } catch(e) {
    document.getElementById('geo-result').innerHTML = ` + "`" + `<span style="color:#f85149">${e.message}</span>` + "`" + `;
  }
}

// ── capacity test ─────────────────────────────────────
let capRunning = false;

async function runCapTest() {
  if (capRunning) return;
  capRunning = true;
  document.getElementById('cap-result').innerHTML = '<span style="color:#8b949e">running internal benchmark…</span>';
  document.getElementById('cap-progress').style.width = '10%';

  // 1. Internal benchmark (server-side)
  const n    = parseInt(document.getElementById('cap-n').value)    || 50000;
  const conc = parseInt(document.getElementById('cap-conc').value) || 20;
  const dur  = parseInt(document.getElementById('cap-dur').value)  || 3;

  let internalRPS = 0, nsPerOp = 0;
  try {
    const res  = await fetch('/bench', {
      method: 'POST', headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({n, cluster: 'api'}),
    });
    const data = await res.json();
    internalRPS = data.rps;
    nsPerOp     = data.ns_per_op;
  } catch(e) { showError('bench failed: ' + e.message); capRunning = false; return; }

  document.getElementById('cap-progress').style.width = '30%';
  document.getElementById('cap-result').innerHTML = '<span style="color:#8b949e">running HTTP benchmark…</span>';

  // 2. HTTP benchmark (browser-side concurrent fetch)
  let httpCount = 0, httpErrors = 0;
  const body    = JSON.stringify({cluster:'api', algo:'round-robin'});
  const headers = {'Content-Type': 'application/json'};
  const deadline = Date.now() + dur * 1000;
  const totalDur = dur * 1000;

  const worker = async () => {
    while (Date.now() < deadline) {
      try {
        const r = await fetch('/route', {method:'POST', headers, body});
        if (r.ok) httpCount++; else httpErrors++;
      } catch(_) { httpErrors++; }
      // update progress bar
      const pct = 30 + 65 * Math.min(1, (dur*1000 - (deadline - Date.now())) / totalDur);
      document.getElementById('cap-progress').style.width = pct + '%';
    }
  };

  await Promise.all(Array.from({length: conc}, worker));
  document.getElementById('cap-progress').style.width = '100%';

  const httpRPS  = Math.round(httpCount / dur);
  const errRate  = httpCount + httpErrors > 0
    ? ((httpErrors / (httpCount + httpErrors)) * 100).toFixed(1)
    : '0.0';

  document.getElementById('cap-result').innerHTML = ` + "`" + `
    <div class="cap-grid">
      <div>
        <div class="cap-big">${internalRPS.toLocaleString()}</div>
        <div class="cap-label">internal req/s</div>
        <div style="font-size:11px;color:#484f58;margin-top:2px">${nsPerOp} ns/op · pure Go routing speed</div>
      </div>
      <div>
        <div class="cap-big">${httpRPS.toLocaleString()}</div>
        <div class="cap-label">HTTP req/s</div>
        <div style="font-size:11px;color:#484f58;margin-top:2px">${conc} concurrent · ${httpErrors} errors (${errRate}%)</div>
      </div>
    </div>
    <div class="cap-note">
      Internal = raw routing speed (mutex + map lookup, no network).<br>
      HTTP = real end-to-end: TCP, HTTP/1.1 keep-alive, JSON encode/decode, mux dispatch.<br>
      The gap between the two is the HTTP stack overhead per request.
    </div>` + "`" + `;
  capRunning = false;
}

// ── failover lab ─────────────────────────────────────
const events = [];

function logEvent(msg, cls) {
  const time = new Date().toLocaleTimeString();
  events.unshift({time, msg, cls});
  if (events.length > 20) events.pop();
  renderEvents();
}

function renderEvents() {
  const ul = document.getElementById('event-list');
  if (!events.length) { ul.innerHTML = '<li style="color:#484f58">no events yet</li>'; return; }
  ul.innerHTML = events.map(e =>
    ` + "`" + `<li><span class="ev-time">${e.time}</span><span class="${e.cls}">${e.msg}</span></li>` + "`" + `
  ).join('');
}

async function dcAction(cluster, dc, action) {
  try {
    const res = await fetch('/failover', {
      method: 'POST', headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({cluster, dc, action}),
    });
    if (!res.ok) { const e = await res.json(); throw new Error(e.error); }
    const data = await res.json();
    const cls  = action === 'unhealthy' ? 'ev-fail' : action === 'draining' ? 'ev-drain' : 'ev-ok';
    const verb = action === 'unhealthy' ? '⚡ failed' : action === 'draining' ? '⏳ draining' : '✓ recovered';
    logEvent(` + "`" + `${dc} DC ${verb} — ${data.affected} backend(s) → ${action}` + "`" + `, cls);
    load();
  } catch(e) { showError(e.message); }
}

function renderDCGrid(clusters) {
  // aggregate backends by dc across all clusters
  const dcs = {};
  for (const c of (clusters || [])) {
    for (const b of Object.values(c.backends || {})) {
      if (!dcs[b.dc]) dcs[b.dc] = {healthy: 0, unhealthy: 0, draining: 0, cluster: c.name};
      dcs[b.dc][b.health]++;
    }
  }
  const grid = document.getElementById('dc-grid');
  const entries = Object.entries(dcs);
  if (!entries.length) { grid.innerHTML = '<p style="color:#8b949e;font-size:12px">no backends registered</p>'; return; }

  grid.innerHTML = entries.map(([dc, s]) => {
    const total   = s.healthy + s.unhealthy + s.draining;
    const impaired = s.unhealthy > 0;
    const draining = s.draining > 0 && s.unhealthy === 0;
    const cardCls  = impaired ? 'dc-card dc-impaired' : draining ? 'dc-card dc-draining' : 'dc-card';
    const countParts = [];
    if (s.healthy)   countParts.push(` + "`" + `<span class="ok">${s.healthy} healthy</span>` + "`" + `);
    if (s.draining)  countParts.push(` + "`" + `<span class="drain">${s.draining} draining</span>` + "`" + `);
    if (s.unhealthy) countParts.push(` + "`" + `<span class="bad">${s.unhealthy} unhealthy</span>` + "`" + `);
    return ` + "`" + `
    <div class="${cardCls}">
      <div class="dc-name">${dc}</div>
      <div class="dc-count">${countParts.join(' · ')} / ${total} total</div>
      <div class="dc-actions">
        <button class="btn btn-red" onclick="dcAction('${s.cluster}','${dc}','unhealthy')"
          data-tip="⚡ Sudden failure&#10;&#10;All ${dc} backends go unhealthy instantly — simulates a DC going dark (power outage, network partition). The router stops sending traffic here on the very next request.&#10;&#10;Watch the dispatcher bars: ${dc}'s bar will flatline and the other DC absorbs 100% of traffic.">⚡ Fail</button>
        <button class="btn btn-yellow" onclick="dcAction('${s.cluster}','${dc}','draining')"
          data-tip="⏳ Graceful drain&#10;&#10;All ${dc} backends enter draining state — simulates a planned maintenance window. No NEW requests are routed here, but any in-flight requests are allowed to finish.&#10;&#10;Safer than a sudden fail for deploys and upgrades.">⏳ Drain</button>
        <button class="btn btn-green" onclick="dcAction('${s.cluster}','${dc}','healthy')"
          data-tip="✓ Recover DC&#10;&#10;All ${dc} backends return to healthy — simulates the DC coming back online after an outage or finishing maintenance.&#10;&#10;The router will immediately start sending traffic here again. Watch the bars rebalance.">✓ Recover</button>
      </div>
    </div>` + "`" + `;
  }).join('');
}

load();
setInterval(load, 2000);
setInterval(loadDispStats, 500);
</script>
</body>
</html>`
