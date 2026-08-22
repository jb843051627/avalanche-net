async function getJSON(url) {
  const resp = await fetch(url);
  if (!resp.ok) throw new Error(url + " -> " + resp.status);
  return resp.json();
}

function badge(text, cls) {
  return '<span class="badge ' + cls + '">' + text + "</span>";
}

async function loadOverview() {
  try {
    const ov = await getJSON("/api/v1/overview");
    document.getElementById("overview").textContent =
      "站点 " + ov.stations + " · 在线 " + ov.online + " · 活跃 critical 告警 " + ov.active_critical;
  } catch (e) { /* 忽略首屏失败 */ }
}

async function loadStations() {
  const tbody = document.querySelector("#station-table tbody");
  try {
    const stations = await getJSON("/api/v1/stations");
    tbody.innerHTML = stations.map(function (s) {
      return "<tr><td>" + s.id + "</td><td>" + s.name + "</td><td>" + s.region_id +
        "</td><td>" + s.elevation_m + " m</td><td>" + s.aspect + "</td><td>" +
        badge(s.status, s.status) + "</td><td>" + (s.last_heartbeat || "-") + "</td></tr>";
    }).join("");
  } catch (e) {
    tbody.innerHTML = '<tr><td colspan="7">加载失败: ' + e.message + "</td></tr>";
  }
}

async function loadAlerts() {
  const tbody = document.querySelector("#alert-table tbody");
  try {
    const stations = await getJSON("/api/v1/stations");
    let rows = [];
    for (const st of stations) {
      const alerts = await getJSON("/api/v1/stations/" + st.id + "/alerts");
      rows = rows.concat(alerts.slice(0, 5));
    }
    rows.sort(function (a, b) { return b.triggered_at < a.triggered_at ? -1 : 1; });
    tbody.innerHTML = rows.slice(0, 15).map(function (a) {
      return "<tr><td>" + a.station_id + "</td><td>" + a.rule_key + "</td><td>" +
        badge(a.level, a.level) + "</td><td>" + a.reason + "</td><td>" +
        a.triggered_at + "</td></tr>";
    }).join("") || '<tr><td colspan="5">暂无活跃告警</td></tr>';
  } catch (e) {
    tbody.innerHTML = '<tr><td colspan="5">加载失败: ' + e.message + "</td></tr>";
  }
}

async function loadBulletins() {
  const tbody = document.querySelector("#bulletin-table tbody");
  try {
    const regions = await getJSON("/api/v1/regions");
    let rows = [];
    for (const r of regions) {
      const list = await getJSON("/api/v1/bulletins?region=" + encodeURIComponent(r[0]));
      if (list.length > 0) rows.push(list[0]);
    }
    tbody.innerHTML = rows.map(function (b) {
      return "<tr><td>" + b.region_id + "</td><td>" + b.issued_for.slice(0, 10) + "</td><td>" +
        b.above_treeline + "</td><td>" + b.near_treeline + "</td><td>" + b.below_treeline +
        "</td><td>" + b.stage + "</td></tr>";
    }).join("") || '<tr><td colspan="6">暂无公报</td></tr>';
  } catch (e) {
    tbody.innerHTML = '<tr><td colspan="6">加载失败: ' + e.message + "</td></tr>";
  }
}

function refreshAll() {
  loadOverview();
  loadStations();
  loadAlerts();
  loadBulletins();
}
refreshAll();
setInterval(refreshAll, 30000);
