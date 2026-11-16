package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Home serves a small server-rendered dashboard. It is the only page in the
// template: it exists so the running service can be seen, not just curled.
func Home(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(homePage))
}

const homePage = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Items · go-service-template</title>
<style>
  :root { --ink:#0f172a; --muted:#64748b; --line:#e2e8f0; --bg:#f6f8fb; --accent:#0f766e; }
  * { box-sizing: border-box; }
  body { margin:0; background:var(--bg); color:var(--ink);
         font:14px/1.55 -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; }
  header { background:#fff; border-bottom:1px solid var(--line); padding:0 32px; height:64px;
           display:flex; align-items:center; justify-content:space-between; }
  .brand { display:flex; align-items:center; gap:12px; }
  .logo { width:28px; height:28px; border-radius:8px; background:#0f766e; color:#fff;
          display:grid; place-items:center; font-weight:700; font-size:13px; }
  h1 { font-size:15px; margin:0; font-weight:650; }
  .sub { font-size:11.5px; color:var(--muted); }
  .pill { font-size:11.5px; color:#0f766e; background:#ecfdf5; border:1px solid #ccf0e4;
          border-radius:999px; padding:4px 10px; }
  main { max-width:980px; margin:24px auto; padding:0 24px; display:grid; gap:18px; }
  .card { background:#fff; border:1px solid var(--line); border-radius:14px; padding:18px 20px; }
  .card h2 { font-size:11px; letter-spacing:.1em; text-transform:uppercase; color:var(--muted);
             margin:0 0 14px; font-weight:600; }
  form { display:grid; grid-template-columns: 1.1fr 1.8fr 140px 110px; gap:10px; }
  input, select { font:inherit; padding:9px 11px; border:1px solid var(--line); border-radius:9px;
                  background:#fff; color:var(--ink); }
  input:focus, select:focus { outline:2px solid #99f6e4; border-color:#5eead4; }
  button { font:inherit; font-weight:600; border:0; border-radius:9px; background:var(--accent);
           color:#fff; padding:9px 16px; cursor:pointer; }
  button.ghost { background:#fff; color:#b91c1c; border:1px solid #fecaca; padding:4px 10px;
                 font-size:12px; font-weight:500; }
  table { width:100%; border-collapse:collapse; }
  th { text-align:left; font-size:11px; letter-spacing:.06em; text-transform:uppercase;
       color:var(--muted); font-weight:600; padding:8px 10px; border-bottom:1px solid var(--line); }
  td { padding:11px 10px; border-bottom:1px solid #f1f5f9; vertical-align:top; }
  td.id { color:var(--muted); font-variant-numeric:tabular-nums; width:52px; }
  .status { font-size:11px; padding:2px 8px; border-radius:999px; background:#ecfdf5; color:#0f766e; }
  .status.archived { background:#f1f5f9; color:#64748b; }
  .desc { color:var(--muted); font-size:12.5px; }
  .empty { color:var(--muted); padding:18px 0; text-align:center; }
  .err { color:#b91c1c; font-size:12.5px; min-height:18px; }
  .features { display:grid; grid-template-columns:repeat(3, 1fr); gap:14px; }
  .feature { background:#fff; border:1px solid var(--line); border-radius:12px; padding:14px 16px;
             display:flex; flex-direction:column; gap:6px; }
  .feature b { font-size:12.5px; }
  .feature span { font-size:12px; color:var(--muted); line-height:1.5; }
  footer { max-width:980px; margin:0 auto 32px; padding:0 24px; color:var(--muted); font-size:12px; }
  code { font:12px ui-monospace, Menlo, monospace; background:#f1f5f9; padding:1px 5px; border-radius:4px; }
  @media (max-width: 760px) { form { grid-template-columns:1fr; } }
</style>
</head>
<body>
<header>
  <div class="brand"><div class="logo">GT</div>
    <div><h1>Items</h1><div class="sub">go-service-template · Gin + GORM + SQLite</div></div>
  </div>
  <span class="pill" id="health">checking…</span>
</header>
<main>
  <section class="card">
    <h2>Create an item</h2>
    <form id="create-form">
      <input name="name" placeholder="Name (required)" maxlength="120" required>
      <input name="description" placeholder="Description" maxlength="2000">
      <select name="status"><option value="active">active</option><option value="archived">archived</option></select>
      <button type="submit">Create</button>
    </form>
    <div class="err" id="error"></div>
  </section>

  <section class="card">
    <h2>Items <span id="count" class="sub"></span></h2>
    <table>
      <thead><tr><th>#</th><th>Name</th><th>Description</th><th>Status</th><th></th></tr></thead>
      <tbody id="rows"><tr><td colspan="5" class="empty">Loading…</td></tr></tbody>
    </table>
  </section>

  <section class="features">
    <div class="feature"><b>One envelope</b><span>Every route answers {data, error} — clients never branch on shape.</span></div>
    <div class="feature"><b>Correlated logs</b><span>X-Request-ID on the response and in every log line for that request.</span></div>
    <div class="feature"><b>Drains on SIGTERM</b><span>In-flight requests finish before the process exits.</span></div>
  </section>
</main>
<footer>
  API: <code>GET /healthz</code> · <code>GET|POST /api/v1/items</code> ·
  <code>GET|PUT|DELETE /api/v1/items/:id</code>. Responses are wrapped in a
  <code>{data, error}</code> envelope; every response carries <code>X-Request-ID</code>.
</footer>
<script>
const rows = document.getElementById('rows');
const count = document.getElementById('count');
const errorBox = document.getElementById('error');

function escapeHtml(value) {
  return String(value ?? '').replace(/[&<>"']/g, (c) =>
    ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));
}

async function refresh() {
  const response = await fetch('/api/v1/items?limit=50');
  const body = await response.json();
  if (!response.ok) { errorBox.textContent = body.error?.message || 'load failed'; return; }
  const items = body.data.items || [];
  count.textContent = '(' + body.data.total + ')';
  rows.innerHTML = items.length ? items.map((item) => ` + "`" + `
    <tr>
      <td class="id">${item.id}</td>
      <td><strong>${escapeHtml(item.name)}</strong></td>
      <td class="desc">${escapeHtml(item.description) || '—'}</td>
      <td><span class="status ${item.status === 'archived' ? 'archived' : ''}">${item.status}</span></td>
      <td><button class="ghost" data-id="${item.id}" type="button">Delete</button></td>
    </tr>` + "`" + `).join('') : '<tr><td colspan="5" class="empty">No items yet.</td></tr>';

  rows.querySelectorAll('button[data-id]').forEach((button) => {
    button.addEventListener('click', async () => {
      await fetch('/api/v1/items/' + button.dataset.id, { method: 'DELETE' });
      refresh();
    });
  });
}

document.getElementById('create-form').addEventListener('submit', async (event) => {
  event.preventDefault();
  errorBox.textContent = '';
  const data = Object.fromEntries(new FormData(event.target).entries());
  const response = await fetch('/api/v1/items', {
    method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(data),
  });
  const body = await response.json();
  if (!response.ok) {
    const fields = body.error?.fields || {};
    errorBox.textContent = Object.entries(fields).map(([k, v]) => k + ' ' + v).join(' · ') ||
      body.error?.message || 'create failed';
    return;
  }
  event.target.reset();
  refresh();
});

fetch('/healthz').then((r) => r.json()).then((body) => {
  const health = body.data || {};
  document.getElementById('health').textContent = health.status + ' · v' + health.version;
}).catch(() => { document.getElementById('health').textContent = 'unreachable'; });

refresh();
</script>
</body>
</html>
`
