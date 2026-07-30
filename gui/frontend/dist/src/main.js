/* After Photo GUI — 前端逻辑
 * 后端方法：window.go.main.App.*
 * 后端事件：log / progress / step / confirm / done
 */

'use strict';

// ---------- 步骤配置（与后端 stepDefs 顺序一致） ----------

const STEPS = [
  { name: '类型拆分', desc: '按 JPG / RAW / 视频归类到子目录', danger: false, checked: true },
  { name: '重复检测', desc: 'pHash 感知哈希 + 时间阈值分组', danger: false, checked: true },
  { name: '智能选优', desc: '清晰度与亮度评分，标记最佳照片', danger: false, checked: true },
  { name: 'RAW 清理', desc: '删除无对应 JPG 的 RAW 文件', danger: true, checked: false },
];

// ---------- DOM ----------

const $ = (id) => document.getElementById(id);
const els = {
  dirPath: $('dir-path'), dirError: $('dir-error'), dirStats: $('dir-stats'),
  statJpg: $('stat-jpg'), statRaw: $('stat-raw'), statVideo: $('stat-video'), statSize: $('stat-size'),
  steps: $('steps'),
  btnBrowse: $('btn-browse'), btnRun: $('btn-run'), btnCancel: $('btn-cancel'),
  pill: $('status-pill'), fill: $('progress-fill'), progressText: $('progress-text'),
  elapsed: $('elapsed'), console: $('console'), statusbar: $('statusbar'),
  btnScroll: $('btn-scroll'), btnClear: $('btn-clear'), btnTheme: $('btn-theme'),
  modalMask: $('modal-mask'), modalMsg: $('modal-msg'),
  btnModalYes: $('btn-modal-yes'), btnModalNo: $('btn-modal-no'),
};

// ---------- 状态 ----------

const state = {
  dir: '',
  running: false,
  autoScroll: true,
  timerId: null,
  startAt: 0,
};

// ---------- 步骤卡片 ----------

const ICON_SPIN = '<svg class="spin" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M21 12a9 9 0 1 1-6.2-8.56"/></svg>';
const ICON_CHECK = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="20 6 9 17 4 12"/></svg>';
const ICON_DOT = '<svg viewBox="0 0 24 24" fill="currentColor" stroke="none"><circle cx="12" cy="12" r="3.5"/></svg>';

function buildSteps() {
  els.steps.innerHTML = '';
  STEPS.forEach((s, i) => {
    const card = document.createElement('div');
    card.className = 'step' + (s.danger ? ' danger' : '') + (s.checked ? ' on' : '');
    card.dataset.index = i;
    card.innerHTML = `
      <span class="step-num">${i + 1}</span>
      <div class="step-info">
        <span class="step-name">${s.name}</span>
        <span class="step-desc">${s.desc}</span>
      </div>
      <span class="step-state">${ICON_DOT}</span>`;
    card.addEventListener('click', () => {
      if (state.running) return;
      s.checked = !s.checked;
      card.classList.toggle('on', s.checked);
    });
    els.steps.appendChild(card);
  });
}

function setStepState(index, cls) {
  const card = els.steps.children[index];
  if (!card) return;
  card.classList.remove('running', 'finished');
  const icon = card.querySelector('.step-state');
  if (cls === 'running') { card.classList.add('running'); icon.innerHTML = ICON_SPIN; }
  else if (cls === 'finished') { card.classList.add('finished'); icon.innerHTML = ICON_CHECK; }
  else icon.innerHTML = ICON_DOT;
}

function resetStepStates() {
  for (let i = 0; i < STEPS.length; i++) setStepState(i, null);
}

// ---------- ANSI 渲染 ----------

const FG = { 31: 'fg-red', 32: 'fg-green', 33: 'fg-yellow', 34: 'fg-blue', 35: 'fg-magenta', 36: 'fg-cyan', 37: 'fg-white' };
const ansiState = { fg: null, bold: false };
let ansiPending = '';

function escHtml(s) {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
}

function ansiToHtml(text) {
  text = ansiPending + text;
  ansiPending = '';
  let html = '';
  let i = 0;
  const openSpan = () => {
    const cls = [ansiState.fg, ansiState.bold ? 'bold' : null].filter(Boolean).join(' ');
    return cls ? `<span class="${cls}">` : '';
  };
  while (i < text.length) {
    const esc = text.indexOf('\x1b[', i);
    if (esc === -1) break;
    if (esc > i) html += openSpan() + escHtml(text.slice(i, esc)) + (openSpan() ? '</span>' : '');
    let j = esc + 2;
    while (j < text.length && !(text.charCodeAt(j) >= 0x40 && text.charCodeAt(j) <= 0x7e)) j++;
    if (j >= text.length) { ansiPending = text.slice(esc); return html; } // 转义序列跨块，留存
    const code = text[j];
    if (code === 'm') {
      for (const n of text.slice(esc + 2, j).split(';')) {
        const v = parseInt(n, 10) || 0;
        if (v === 0) { ansiState.fg = null; ansiState.bold = false; }
        else if (v === 1) ansiState.bold = true;
        else if (FG[v]) ansiState.fg = FG[v];
        else if (v === 39) ansiState.fg = null;
      }
    }
    i = j + 1;
  }
  if (i < text.length) html += openSpan() + escHtml(text.slice(i)) + (openSpan() ? '</span>' : '');
  return html;
}

function appendLog(text) {
  els.console.insertAdjacentHTML('beforeend', ansiToHtml(text));
  // 防止超长日志拖垮 DOM
  while (els.console.childNodes.length > 4000) {
    els.console.removeChild(els.console.firstChild);
  }
  if (state.autoScroll) els.console.scrollTop = els.console.scrollHeight;
}

// ---------- 状态展示 ----------

const PILLS = {
  idle: ['pill-idle', '就绪'],
  running: ['pill-running', '处理中'],
  done: ['pill-done', '完成'],
  error: ['pill-error', '已取消'],
};

function setPill(kind) {
  const [cls, label] = PILLS[kind];
  els.pill.className = 'pill ' + cls;
  els.pill.textContent = label;
}

function setProgress(cur, total, msg) {
  if (total > 0) {
    els.fill.classList.remove('indeterminate');
    els.fill.style.width = Math.min(100, (cur / total) * 100) + '%';
    els.progressText.textContent = `${cur}/${total}  ${msg || ''}`;
  } else {
    els.fill.classList.add('indeterminate');
    els.progressText.textContent = msg || '';
  }
}

function resetProgress() {
  els.fill.classList.remove('indeterminate');
  els.fill.style.width = '0%';
  els.progressText.textContent = '';
}

function fmtElapsed(ms) {
  const s = Math.floor(ms / 1000);
  const m = Math.floor(s / 60);
  return m > 0 ? `${m}m ${s % 60}s` : `${s}s`;
}

function fmtSize(bytes) {
  if (bytes >= 1 << 30) return (bytes / (1 << 30)).toFixed(1) + 'G';
  if (bytes >= 1 << 20) return (bytes / (1 << 20)).toFixed(1) + 'M';
  if (bytes >= 1 << 10) return (bytes / (1 << 10)).toFixed(0) + 'K';
  return bytes + 'B';
}

// ---------- 目录 ----------

async function pickDirectory() {
  const dir = await window.go.main.App.PickDirectory();
  if (dir) await scan(dir);
}

async function scan(dir) {
  els.dirError.classList.add('hidden');
  try {
    const stats = await window.go.main.App.ScanDirectory(dir);
    state.dir = dir;
    els.dirPath.textContent = dir;
    els.dirPath.classList.remove('empty');
    els.dirPath.title = dir;
    els.statJpg.textContent = stats.jpg;
    els.statRaw.textContent = stats.raw;
    els.statVideo.textContent = stats.video;
    els.statSize.textContent = fmtSize(stats.totalSize);
    els.dirStats.classList.remove('hidden');
    els.btnRun.disabled = false;
    els.statusbar.textContent = '目录已就绪，点击「开始处理」执行';
  } catch (err) {
    state.dir = '';
    els.btnRun.disabled = true;
    els.dirStats.classList.add('hidden');
    els.dirError.textContent = String(err);
    els.dirError.classList.remove('hidden');
  }
}

// ---------- 运行控制 ----------

function setRunningUI(running) {
  state.running = running;
  els.btnRun.classList.toggle('hidden', running);
  els.btnCancel.classList.toggle('hidden', !running);
  els.btnBrowse.disabled = running;
  for (const card of els.steps.children) card.classList.toggle('disabled', running);
}

async function run() {
  if (!state.dir || state.running) return;
  const steps = STEPS.map((s) => s.checked);
  if (!steps.some(Boolean)) {
    els.statusbar.textContent = '请至少勾选一个步骤';
    return;
  }

  setRunningUI(true);
  setPill('running');
  resetStepStates();
  resetProgress();
  els.statusbar.textContent = '任务启动中…';
  state.startAt = Date.now();
  els.elapsed.textContent = '0s';
  state.timerId = setInterval(() => {
    els.elapsed.textContent = fmtElapsed(Date.now() - state.startAt);
  }, 500);

  try {
    await window.go.main.App.StartProcessing(state.dir, steps);
  } catch (err) {
    appendLog(`\n✗ 启动失败：${err}\n`);
    finishUI(true, String(err));
  }
}

function finishUI(cancelled, detail) {
  clearInterval(state.timerId);
  state.timerId = null;
  setRunningUI(false);
  setPill(cancelled ? 'error' : 'done');
  for (let i = 0; i < STEPS.length; i++) {
    const card = els.steps.children[i];
    if (card && card.classList.contains('running')) setStepState(i, 'finished');
  }
  els.fill.classList.remove('indeterminate');
  if (!cancelled) els.fill.style.width = '100%';
  els.statusbar.textContent = cancelled
    ? '任务已取消'
    : `全部完成，用时 ${detail || fmtElapsed(Date.now() - state.startAt)}`;
}

// ---------- 确认弹窗 ----------

function showConfirm(message) {
  els.modalMsg.textContent = message;
  els.modalMask.classList.remove('hidden');
}

function answerConfirm(ok) {
  els.modalMask.classList.add('hidden');
  window.go.main.App.ConfirmStep4(ok);
}

// ---------- 主题 ----------

function toggleTheme() {
  const html = document.documentElement;
  const next = html.dataset.theme === 'dark' ? 'light' : 'dark';
  html.dataset.theme = next;
  try { localStorage.setItem('after-photo-theme', next); } catch (_) {}
}

function loadTheme() {
  try {
    const t = localStorage.getItem('after-photo-theme');
    if (t) document.documentElement.dataset.theme = t;
  } catch (_) {}
}

// ---------- 事件订阅 ----------

function subscribe() {
  window.runtime.EventsOn('log', (text) => appendLog(text));
  window.runtime.EventsOn('progress', (p) => setProgress(p.current, p.total, p.message));
  window.runtime.EventsOn('step', (s) => {
    for (let i = 0; i < STEPS.length; i++) {
      const card = els.steps.children[i];
      if (card && card.classList.contains('running')) setStepState(i, 'finished');
    }
    setStepState(s.index, 'running');
    resetProgress();
    els.statusbar.textContent = s.name;
  });
  window.runtime.EventsOn('confirm', (msg) => showConfirm(msg));
  window.runtime.EventsOn('done', (d) => finishUI(d.cancelled, d.duration));
}

// ---------- 初始化 ----------

function init() {
  loadTheme();
  buildSteps();
  subscribe();

  els.btnBrowse.addEventListener('click', pickDirectory);
  els.btnRun.addEventListener('click', run);
  els.btnCancel.addEventListener('click', () => {
    els.btnCancel.disabled = true;
    els.statusbar.textContent = '正在取消，等待当前步骤结束…';
    window.go.main.App.CancelProcessing();
    setTimeout(() => { els.btnCancel.disabled = false; }, 1500);
  });
  els.btnClear.addEventListener('click', () => { els.console.innerHTML = ''; });
  els.btnScroll.addEventListener('click', () => {
    state.autoScroll = !state.autoScroll;
    els.btnScroll.classList.toggle('active', state.autoScroll);
    if (state.autoScroll) els.console.scrollTop = els.console.scrollHeight;
  });
  els.btnTheme.addEventListener('click', toggleTheme);
  els.btnModalYes.addEventListener('click', () => answerConfirm(true));
  els.btnModalNo.addEventListener('click', () => answerConfirm(false));
}

window.addEventListener('DOMContentLoaded', init);
