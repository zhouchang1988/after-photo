const themes = [
  { id: 'tokyo-night', name: 'Tokyo Night', type: 'dark', bg: '#1a1b26', ink: '#a9b1d6', accent: '#7aa2f7', success: '#9ece6a', danger: '#f7768e' },
  { id: 'dracula', name: 'Dracula', type: 'dark', bg: '#282a36', ink: '#f8f8f2', accent: '#ff79c6', success: '#50fa7b', danger: '#ff5555' },
  { id: 'catppuccin-mocha', name: 'Catppuccin Mocha', type: 'dark', bg: '#1e1e2e', ink: '#cdd6f4', accent: '#cba6f7', success: '#a6e3a1', danger: '#f38ba8' },
  { id: 'catppuccin-macchiato', name: 'Catppuccin Macchiato', type: 'dark', bg: '#181825', ink: '#cad3f8', accent: '#c7a4f5', success: '#a6da95', danger: '#ed8796' },
  { id: 'nord', name: 'Nord', type: 'dark', bg: '#2e3440', ink: '#d8dee9', accent: '#88c0d0', success: '#a3be8c', danger: '#bf616a' },
  { id: 'one-dark-pro', name: 'One Dark Pro', type: 'dark', bg: '#282c34', ink: '#abb2bf', accent: '#4d78cc', success: '#98c379', danger: '#e06c75' },
  { id: 'gruvbox-dark', name: 'Gruvbox Dark', type: 'dark', bg: '#282828', ink: '#ebdbb2', accent: '#fe8019', success: '#b8bb26', danger: '#fb4934' },
  { id: 'kanagawa', name: 'Kanagawa Wave', type: 'dark', bg: '#1f1f28', ink: '#dcd7ba', accent: '#658594', success: '#76956a', danger: '#c34043' },
  { id: 'rose-pine', name: 'Rose Pine', type: 'dark', bg: '#191724', ink: '#e0def4', accent: '#31748f', success: '#9ccfd8', danger: '#eb6f92' },
  { id: 'github-dark', name: 'GitHub Dark', type: 'dark', bg: '#0d1117', ink: '#e6edf3', accent: '#1f6feb', success: '#3fb950', danger: '#f85149' },
  { id: 'material-palenight', name: 'Material Palenight', type: 'dark', bg: '#292d3e', ink: '#eeffff', accent: '#80cbc4', success: '#c3e88d', danger: '#ff5370' },
  { id: 'ayu-dark', name: 'Ayu Dark', type: 'dark', bg: '#0b0e14', ink: '#bfbdb6', accent: '#e6b450', success: '#c2d94c', danger: '#f07178' },
  { id: 'vitesse-dark', name: 'Vitesse Dark', type: 'dark', bg: '#121212', ink: '#dbd7ca', accent: '#4d9375', success: '#80a665', danger: '#cb7676' },
  { id: 'buddy-dark', name: 'Default Dark', type: 'dark', bg: '#18181a', ink: '#e8e8e3', accent: '#339cff', success: '#40c977', danger: '#fa423e' },
  { id: 'codex-dark', name: 'Codex Dark', type: 'dark', bg: '#111111', ink: '#ffffff', accent: '#0169cc', success: '#40c977', danger: '#fa423e' },
  { id: 'atom-material', name: 'Atom Material', type: 'dark', bg: '#263238', ink: '#eeffff', accent: '#82aaff', success: '#c3e88d', danger: '#f07178' },
  { id: 'gothic', name: 'Gothic', type: 'dark', bg: '#0e0e0e', ink: '#c7c7c7', accent: '#fe5e3a', success: '#40c977', danger: '#b33b2e' },
  { id: 'monokai', name: 'Monokai', type: 'dark', bg: '#282828', ink: '#f8f8f2', accent: '#a6e22e', success: '#a6e22e', danger: '#f92672' },
  { id: 'slack-dark', name: 'Slack Dark', type: 'dark', bg: '#1a1d21', ink: '#d1d2d3', accent: '#1264a3', success: '#2bac76', danger: '#e01e5a' },
  { id: 'dark-plus', name: 'Dark+', type: 'dark', bg: '#1e1e1e', ink: '#d4d4d4', accent: '#569cd6', success: '#6a9955', danger: '#f44747' },
  { id: 'min-dark', name: 'Minimal Dark', type: 'dark', bg: '#1a1a1a', ink: '#e0e0e0', accent: '#4da6ff', success: '#4caf50', danger: '#f44336' },
  { id: 'catppuccin-latte', name: 'Catppuccin Latte', type: 'light', bg: '#eff1f5', ink: '#4c4f69', accent: '#8839ef', success: '#40a02b', danger: '#d20f39' },
  { id: 'github-light', name: 'GitHub Light', type: 'light', bg: '#ffffff', ink: '#1f2328', accent: '#0969da', success: '#1a7f37', danger: '#d1242f' },
  { id: 'one-light', name: 'One Light', type: 'light', bg: '#fafafa', ink: '#383a42', accent: '#4078f2', success: '#50a14f', danger: '#e45649' },
  { id: 'solarized-light', name: 'Solarized Light', type: 'light', bg: '#fdf6e3', ink: '#657b83', accent: '#268bd2', success: '#859900', danger: '#dc322f' },
  { id: 'min-light', name: 'Minimal Light', type: 'light', bg: '#ffffff', ink: '#1a1a1a', accent: '#0066cc', success: '#22863a', danger: '#cb2431' },
  { id: 'slack-light', name: 'Slack Light', type: 'light', bg: '#ffffff', ink: '#1d1c1d', accent: '#1264a3', success: '#2bac76', danger: '#e01e5a' },
  { id: 'light-plus', name: 'Light+', type: 'light', bg: '#ffffff', ink: '#333333', accent: '#0066cc', success: '#22863a', danger: '#cb2431' },
  { id: 'buddy-light', name: 'Default Light', type: 'light', bg: '#ffffff', ink: '#1a1a1a', accent: '#0066cc', success: '#28a745', danger: '#dc3545' },
  { id: 'vitesse-light', name: 'Vitesse Light', type: 'light', bg: '#f7f7f7', ink: '#1a1a1a', accent: '#2993a6', success: '#4e8c2f', danger: '#cb4a5a' },
  { id: 'mpe-atom-light', name: 'Atom Light', type: 'light', bg: '#fafafa', ink: '#383a42', accent: '#4078f2', success: '#50a14f', danger: '#e45649' },
  { id: 'mpe-one-light', name: 'One Light (MPE)', type: 'light', bg: '#fafafa', ink: '#383a42', accent: '#4078f2', success: '#50a14f', danger: '#e45649' },
];

let currentTheme = localStorage.getItem('theme') || 'tokyo-night';
let timerInterval = null;
let startTime = null;
let isProcessing = false;

document.addEventListener('DOMContentLoaded', () => {
  initTheme();
  initThemePanel();
  initEventListeners();
  initWailsEvents();
});

function initTheme() {
  applyTheme(currentTheme);
  renderThemeGrid();
}

function applyTheme(themeId) {
  document.documentElement.setAttribute('data-theme', themeId);
  currentTheme = themeId;
  localStorage.setItem('theme', themeId);
  document.querySelectorAll('.theme-card').forEach(card => {
    card.classList.toggle('active', card.dataset.id === themeId);
  });
}

function renderThemeGrid(filter = 'all') {
  const grid = document.getElementById('theme-grid');
  grid.innerHTML = '';
  const filteredThemes = filter === 'all' ? themes : themes.filter(t => t.type === filter);
  
  filteredThemes.forEach(theme => {
    const card = document.createElement('div');
    card.className = `theme-card ${theme.id === currentTheme ? 'active' : ''}`;
    card.dataset.id = theme.id;
    card.dataset.type = theme.type;
    card.innerHTML = `
      <div class="theme-preview" style="background-color: ${theme.bg}; color: ${theme.ink};">
        <div class="theme-preview-header">
          <span class="theme-preview-dot" style="background: ${theme.danger};"></span>
          <span class="theme-preview-dot" style="background: ${theme.success};"></span>
          <span class="theme-preview-dot" style="background: ${theme.accent};"></span>
        </div>
        <div class="theme-preview-body">
          <span class="theme-preview-title" style="color: ${theme.ink};">${theme.name}</span>
          <div class="theme-preview-line" style="background: ${theme.ink}; opacity: 0.3;"></div>
          <div class="theme-preview-line short" style="background: ${theme.ink}; opacity: 0.2;"></div>
          <div class="theme-preview-line" style="background: ${theme.accent}; opacity: 0.4;"></div>
        </div>
      </div>
      <div class="theme-colors">
        <div class="theme-color-swatch" style="background: ${theme.bg};"></div>
        <div class="theme-color-swatch" style="background: ${theme.ink};"></div>
        <div class="theme-color-swatch" style="background: ${theme.accent};"></div>
        <div class="theme-color-swatch" style="background: ${theme.success};"></div>
        <div class="theme-color-swatch" style="background: ${theme.danger};"></div>
      </div>
      <span class="theme-name">${theme.name}</span>
    `;
    card.addEventListener('click', () => applyTheme(theme.id));
    grid.appendChild(card);
  });
}

function initThemePanel() {
  const toggle = document.getElementById('theme-toggle');
  const panel = document.getElementById('theme-panel');
  const close = document.getElementById('theme-close');
  const filterBtns = document.querySelectorAll('.filter-btn');
  
  toggle.addEventListener('click', () => panel.classList.toggle('hidden'));
  close.addEventListener('click', () => panel.classList.add('hidden'));
  
  filterBtns.forEach(btn => {
    btn.addEventListener('click', () => {
      filterBtns.forEach(b => b.classList.remove('active'));
      btn.classList.add('active');
      renderThemeGrid(btn.dataset.filter);
    });
  });
  
  document.addEventListener('click', (e) => {
    if (!panel.contains(e.target) && !toggle.contains(e.target)) {
      panel.classList.add('hidden');
    }
  });
}

function initEventListeners() {
  const dirInput = document.getElementById('dir-input');
  const btnBrowse = document.getElementById('btn-browse');
  const btnStart = document.getElementById('btn-start');
  const btnClear = document.getElementById('btn-clear');
  const btnConfirmYes = document.getElementById('btn-confirm-yes');
  const btnConfirmNo = document.getElementById('btn-confirm-no');
  const helpToggle = document.getElementById('help-toggle');
  const helpClose = document.getElementById('help-close');
  const helpDialog = document.getElementById('help-dialog');
  
  btnBrowse.addEventListener('click', async () => {
    if (window.go && window.go.main && window.go.main.App) {
      const dir = await window.go.main.App.PickDirectory();
      if (dir) {
        dirInput.value = dir;
        validateDir(dir);
      }
    }
  });
  
  dirInput.addEventListener('change', () => {
    const dir = dirInput.value.trim();
    if (dir) validateDir(dir);
  });
  
  dirInput.addEventListener('keypress', (e) => {
    if (e.key === 'Enter') {
      const dir = dirInput.value.trim();
      if (dir) validateDir(dir);
    }
  });
  
  btnStart.addEventListener('click', handleStartProcessing);
  btnClear.addEventListener('click', () => { document.getElementById('output').textContent = ''; });
  btnConfirmYes.addEventListener('click', () => handleConfirm(true));
  btnConfirmNo.addEventListener('click', () => handleConfirm(false));
  
  helpToggle.addEventListener('click', () => helpDialog.classList.remove('hidden'));
  helpClose.addEventListener('click', () => helpDialog.classList.add('hidden'));
  helpDialog.addEventListener('click', (e) => {
    if (e.target === helpDialog) helpDialog.classList.add('hidden');
  });
}

async function validateDir(dir) {
  const dirError = document.getElementById('dir-error');
  const dirStatus = document.getElementById('dir-status');
  const dirPath = document.getElementById('dir-path');
  
  try {
    if (window.go && window.go.main && window.go.main.App) {
      const validDir = await window.go.main.App.ValidateDirectory(dir);
      dirError.classList.add('hidden');
      dirStatus.classList.remove('hidden');
      dirPath.textContent = validDir;
    }
  } catch (err) {
    dirError.textContent = err;
    dirError.classList.remove('hidden');
    dirStatus.classList.add('hidden');
  }
}

function initWailsEvents() {
  const check = setInterval(() => {
    if (window.runtime) {
      clearInterval(check);
      setupEventListeners();
    }
  }, 50);
}

function setupEventListeners() {
  if (!window.runtime) return;
  
  window.runtime.EventsOn('output', (text) => {
    appendOutput(text);
  });
  
  window.runtime.EventsOn('confirm-request', (message) => {
    showConfirmDialog(message);
  });
  
  window.runtime.EventsOn('processing-complete', (duration) => {
    handleProcessingComplete(duration);
  });
}

async function handleStartProcessing() {
  if (isProcessing) return;
  
  const dirInput = document.getElementById('dir-input');
  const dir = dirInput.value.trim();
  
  const steps = [
    document.getElementById('step1').checked,
    document.getElementById('step2').checked,
    document.getElementById('step3').checked,
    document.getElementById('step4').checked,
  ];
  
  if (!steps.some(s => s)) {
    appendOutput('错误: 请至少选择一个步骤\n');
    return;
  }
  
  isProcessing = true;
  document.getElementById('output').textContent = '';
  startTimer();
  
  const statusEl = document.getElementById('status-text');
  statusEl.className = 'status processing';
  statusEl.textContent = '处理中...';
  
  const btnStart = document.getElementById('btn-start');
  btnStart.disabled = true;
  btnStart.style.opacity = '0.6';
  
  try {
    if (window.go && window.go.main && window.go.main.App) {
      await window.go.main.App.StartProcessing(dir, steps);
    }
  } catch (err) {
    appendOutput(`错误: ${err}\n`);
    resetProcessingState();
  }
}

function resetProcessingState() {
  isProcessing = false;
  const btnStart = document.getElementById('btn-start');
  btnStart.disabled = false;
  btnStart.style.opacity = '1';
  stopTimer();
}

function appendOutput(text) {
  const output = document.getElementById('output');
  output.textContent += text;
  output.scrollTop = output.scrollHeight;
}

function startTimer() {
  startTime = Date.now();
  const timerEl = document.getElementById('timer');
  if (timerInterval) clearInterval(timerInterval);
  
  timerInterval = setInterval(() => {
    const elapsed = Date.now() - startTime;
    const minutes = Math.floor(elapsed / 60000);
    const seconds = Math.floor((elapsed % 60000) / 1000);
    timerEl.textContent = `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`;
  }, 1000);
}

function stopTimer() {
  if (timerInterval) {
    clearInterval(timerInterval);
    timerInterval = null;
  }
}

function showConfirmDialog(message) {
  document.getElementById('confirm-message').textContent = message;
  document.getElementById('confirm-dialog').classList.remove('hidden');
}

function hideConfirmDialog() {
  document.getElementById('confirm-dialog').classList.add('hidden');
}

async function handleConfirm(confirmed) {
  hideConfirmDialog();
  try {
    if (window.go && window.go.main && window.go.main.App) {
      const err = await window.go.main.App.ConfirmStep4(confirmed);
      if (err) appendOutput(`确认失败: ${err}\n`);
    }
  } catch (err) {
    appendOutput(`确认错误: ${err}\n`);
  }
}

function handleProcessingComplete(duration) {
  resetProcessingState();
  const statusEl = document.getElementById('status-text');
  statusEl.className = 'status success';
  statusEl.textContent = `完成 ${duration}`;
}
