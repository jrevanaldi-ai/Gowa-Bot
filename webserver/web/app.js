async function fetchJSON(url) {
    try {
        const r = await fetch(url, { cache: 'no-store' });
        if (!r.ok) return null;
        return await r.json();
    } catch (e) {
        return null;
    }
}

function fmtBytes(mb) {
    if (mb < 1) return (mb * 1024).toFixed(1) + ' KB';
    return mb.toFixed(2) + ' MB';
}

function fmtDuration(sec) {
    sec = Math.floor(sec);
    const d = Math.floor(sec / 86400);
    const h = Math.floor((sec % 86400) / 3600);
    const m = Math.floor((sec % 3600) / 60);
    const s = sec % 60;
    if (d) return `${d}d ${h}h ${m}m`;
    if (h) return `${h}h ${m}m ${s}s`;
    if (m) return `${m}m ${s}s`;
    return `${s}s`;
}

function fmtRelative(isoTs) {
    if (!isoTs) return '-';
    const t = new Date(isoTs).getTime();
    if (!t || t <= 0) return '-';
    const diff = Math.floor((Date.now() - t) / 1000);
    if (diff < 0) return 'just now';
    if (diff < 60) return diff + 's ago';
    if (diff < 3600) return Math.floor(diff / 60) + 'm ago';
    if (diff < 86400) return Math.floor(diff / 3600) + 'h ago';
    return Math.floor(diff / 86400) + 'd ago';
}

function shortID(id) {
    if (!id) return '-';
    if (id.length <= 12) return id;
    return id.slice(0, 8) + '…' + id.slice(-4);
}

function kv(k, v) {
    return `<div class="kv"><span class="k">${k}</span><span class="v">${v}</span></div>`;
}

function escapeHTML(s) {
    return String(s).replace(/[&<>"']/g, c => ({
        '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;'
    }[c]));
}

function renderServer(info) {
    if (!info) return '<p class="empty">Tidak bisa load data server</p>';
    return [
        kv('Hostname', escapeHTML(info.hostname || '-')),
        kv('Uptime', fmtDuration(info.uptime_seconds)),
        kv('Started', new Date(info.started_at).toLocaleString('id-ID')),
        kv('Platform', `${info.os}/${info.arch}`),
        kv('Go', escapeHTML(info.go_version || '-')),
        kv('CPUs', info.cpus),
        kv('Goroutines', info.goroutines)
    ].join('');
}

function renderMemory(mem) {
    if (!mem) return '<p class="empty">Tidak bisa load data memory</p>';
    return [
        kv('RSS', fmtBytes(mem.rss_mb)),
        kv('Heap In-Use', fmtBytes(mem.heap_inuse_mb)),
        kv('Heap Idle', fmtBytes(mem.heap_idle_mb)),
        kv('Stack', fmtBytes(mem.stack_inuse_mb)),
        kv('Sys Total', fmtBytes(mem.sys_mb)),
        kv('Total Alloc', fmtBytes(mem.total_alloc_mb)),
        kv('GC Count', mem.num_gc),
        kv('Last GC', fmtRelative(mem.last_gc))
    ].join('');
}

function renderBot(bot) {
    if (!bot) return '<p class="empty">Tidak bisa load data bot</p>';
    const statusVal = bot.connected
        ? '<span class="pill active">Online</span>'
        : (bot.logged_in
            ? '<span class="pill inactive">Disconnected</span>'
            : '<span class="pill inactive">Not paired</span>');
    return [
        kv('Status', statusVal),
        kv('Phone', escapeHTML(bot.phone || '-')),
        kv('JID', `<code class="mono">${escapeHTML(bot.jid || '-')}</code>`),
        kv('Push Name', escapeHTML(bot.push_name || '-')),
        kv('Self Mode', bot.self_mode ? '<span class="pill active">On</span>' : '<span class="pill inactive">Off</span>'),
        kv('Prefixes', (bot.prefixes || []).map(p => `<code>${escapeHTML(p)}</code>`).join(' ') || '-')
    ].join('');
}

function renderJadibots(data) {
    if (!data) return '<p class="empty">Tidak bisa load data jadibot</p>';
    document.getElementById('jadibot-count').textContent = data.total;
    if (!data.items || data.items.length === 0) {
        return '<p class="empty">Belum ada jadibot aktif</p>';
    }
    let html = '<table><thead><tr><th>ID</th><th>Phone</th><th>Owner</th><th>Status</th><th>Running</th></tr></thead><tbody>';
    data.items.forEach(j => {
        const owner = (j.owner_jid || '').split('@')[0] || '-';
        const statusPill = j.status === 'active'
            ? `<span class="pill active">${escapeHTML(j.status)}</span>`
            : `<span class="pill inactive">${escapeHTML(j.status)}</span>`;
        const runningPill = j.running
            ? '<span class="pill active">Yes</span>'
            : '<span class="pill inactive">No</span>';
        html += `<tr>
            <td><code class="mono" title="${escapeHTML(j.id || '')}">${escapeHTML(shortID(j.id))}</code></td>
            <td>${escapeHTML(j.phone_number)}</td>
            <td>${escapeHTML(owner)}</td>
            <td>${statusPill}</td>
            <td>${runningPill}</td>
        </tr>`;
    });
    html += '</tbody></table>';
    return html;
}

function renderCommands(data) {
    if (!data) return '<p class="empty">Tidak bisa load data commands</p>';
    document.getElementById('commands-count').textContent = data.total;
    if (!data.commands || data.commands.length === 0) {
        return '<p class="empty">Belum ada command</p>';
    }
    const byTag = {};
    data.commands.forEach(c => {
        (byTag[c.tag] = byTag[c.tag] || []).push(c);
    });
    const tags = Object.keys(byTag).sort((a, b) => a.localeCompare(b));
    let html = '';
    tags.forEach(tag => {
        const cmds = byTag[tag].sort((a, b) => a.cmd.localeCompare(b.cmd));
        html += `<div class="cmd-group">
            <h3>${escapeHTML(tag)}</h3>
            <div class="cmd-list">
                ${cmds.map(c => {
                    const cls = c.owner_only ? 'cmd-item owner-only' : 'cmd-item';
                    const title = `${c.desc || ''}${c.example ? ' — ' + c.example : ''}`;
                    return `<span class="${cls}" title="${escapeHTML(title)}">.${escapeHTML(c.cmd)}</span>`;
                }).join('')}
            </div>
        </div>`;
    });
    return html;
}

function applyStatus(bot) {
    const statusEl = document.getElementById('bot-status');
    if (!statusEl) return;
    const label = statusEl.querySelector('.label') || statusEl;
    if (bot && bot.connected) {
        statusEl.className = 'status-pill connected';
        label.textContent = 'Connected';
    } else if (bot) {
        statusEl.className = 'status-pill disconnected';
        label.textContent = 'Disconnected';
    } else {
        statusEl.className = 'status-pill';
        label.textContent = 'Unknown';
    }
}

function applyDynamic(payload) {
    if (!payload) return;
    if (payload.info) document.getElementById('server').innerHTML = renderServer(payload.info);
    if (payload.memory) document.getElementById('memory').innerHTML = renderMemory(payload.memory);
    if (payload.bot) {
        document.getElementById('bot').innerHTML = renderBot(payload.bot);
        applyStatus(payload.bot);
    }
    if (payload.jadibots) document.getElementById('jadibots').innerHTML = renderJadibots(payload.jadibots);
    document.getElementById('last-update').textContent =
        'Live · ' + new Date().toLocaleTimeString('id-ID');
}

function applyCommands(commands) {
    if (!commands) return;
    document.getElementById('commands').innerHTML = renderCommands(commands);
}

let es = null;
let reconnectTimer = null;

function connectStream() {
    if (reconnectTimer) {
        clearTimeout(reconnectTimer);
        reconnectTimer = null;
    }
    if (es) {
        try { es.close(); } catch (_) {}
    }
    es = new EventSource('/api/stream');

    es.addEventListener('init', (e) => {
        try {
            const data = JSON.parse(e.data);
            applyDynamic(data);
            applyCommands(data.commands);
        } catch (_) {}
    });

    es.addEventListener('tick', (e) => {
        try { applyDynamic(JSON.parse(e.data)); } catch (_) {}
    });

    es.addEventListener('commands', (e) => {
        try { applyCommands(JSON.parse(e.data)); } catch (_) {}
    });

    es.onerror = () => {
        const lu = document.getElementById('last-update');
        if (lu) lu.textContent = 'Reconnecting…';
        if (reconnectTimer) return;
        reconnectTimer = setTimeout(() => {
            reconnectTimer = null;
            connectStream();
        }, 2000);
    };
}

connectStream();
