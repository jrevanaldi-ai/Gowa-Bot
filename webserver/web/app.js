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
        kv('Started At', new Date(info.started_at).toLocaleString('id-ID')),
        kv('OS', `${info.os}/${info.arch}`),
        kv('Go', info.go_version),
        kv('CPUs', info.cpus),
        kv('Goroutines', info.goroutines),
        kv('PID', info.pid)
    ].join('');
}

function renderMemory(mem) {
    if (!mem) return '<p class="empty">Tidak bisa load data memory</p>';
    return [
        kv('RSS', fmtBytes(mem.rss_mb)),
        kv('Heap Alloc', fmtBytes(mem.heap_alloc_mb)),
        kv('Heap In-Use', fmtBytes(mem.heap_inuse_mb)),
        kv('Heap Idle', fmtBytes(mem.heap_idle_mb)),
        kv('Stack', fmtBytes(mem.stack_inuse_mb)),
        kv('Sys Total', fmtBytes(mem.sys_mb)),
        kv('GC Count', mem.num_gc)
    ].join('');
}

function renderBot(bot) {
    if (!bot) return '<p class="empty">Tidak bisa load data bot</p>';
    return [
        kv('Connected', bot.connected ? '<span class="pill active">Yes</span>' : '<span class="pill inactive">No</span>'),
        kv('Logged In', bot.logged_in ? '<span class="pill active">Yes</span>' : '<span class="pill inactive">No</span>'),
        kv('Phone', escapeHTML(bot.phone || '-')),
        kv('Push Name', escapeHTML(bot.push_name || '-')),
        kv('Self Mode', bot.self_mode ? '<span class="pill active">On</span>' : '<span class="pill inactive">Off</span>'),
        kv('Prefixes', (bot.prefixes || []).map(p => `<code>${escapeHTML(p)}</code>`).join(' '))
    ].join('');
}

function renderJadibots(data) {
    if (!data) return '<p class="empty">Tidak bisa load data jadibot</p>';
    document.getElementById('jadibot-count').textContent = data.total;
    if (!data.items || data.items.length === 0) {
        return '<p class="empty">Belum ada jadibot aktif</p>';
    }
    let html = '<table><thead><tr><th>Phone</th><th>Owner</th><th>Status</th><th>Running</th></tr></thead><tbody>';
    data.items.forEach(j => {
        const owner = (j.owner_jid || '').split('@')[0] || '-';
        const statusPill = j.status === 'active'
            ? `<span class="pill active">${escapeHTML(j.status)}</span>`
            : `<span class="pill inactive">${escapeHTML(j.status)}</span>`;
        const runningPill = j.running
            ? '<span class="pill active">Yes</span>'
            : '<span class="pill inactive">No</span>';
        html += `<tr>
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
