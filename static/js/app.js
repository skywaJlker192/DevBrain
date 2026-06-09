// ===== DevBrain Pro - JavaScript =====
// Требование #3: Все API запросы с токеном и обработкой 204

const state = {
  user: null,
  token: localStorage.getItem('token'),
  links: [],
  tags: [],
  filters: { search: '', tags: [] },
};

const api = {
  async request(endpoint, options = {}) {
    const config = {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
    };

    // Требование #3: Добавляем JWT токен
    if (state.token) {
      config.headers['Authorization'] = `Bearer ${state.token}`;
    }

    try {
      const response = await fetch(`/api${endpoint}`, config);

      // Требование #3: Если 401 — очищаем токен
      if (response.status === 401) {
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        state.token = null;
        state.user = null;
        window.location.reload();
        throw new Error('Unauthorized');
      }

      // Требование #2: Для 204 (No Content) не парсим JSON
      if (response.status === 204) {
        return { success: true };
      }

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.message || 'Ошибка');
      }

      return data;
    } catch (error) {
      if (error.message !== 'Unauthorized') {
        showToast(error.message, 'error');
      }
      throw error;
    }
  },
  get: (url, opts) => api.request(url, { method: 'GET', ...opts }),
  post: (url, body, opts) =>
    api.request(url, { method: 'POST', body: JSON.stringify(body), ...opts }),
  delete: (url, opts) => api.request(url, { method: 'DELETE', ...opts }),
};

// ===== Auth Functions =====

async function login(email, password) {
  try {
    const data = await api.post('/auth/login', { email, password });
    state.token = data.access_token;
    state.user = data.user;
    localStorage.setItem('token', data.access_token);
    localStorage.setItem('user', JSON.stringify(data.user));
    closeAuthModal();
    showToast('Добро пожаловать!', 'success');
    updateUI();
    await loadLinks();
    await loadStats();
  } catch (e) {}
}

async function register(name, email, password) {
  try {
    await api.post('/auth/register', { name, email, password });
    showToast('Регистрация успешна! Теперь войдите.', 'success');
    switchAuthTab('login');
  } catch (e) {}
}

function logout() {
  state.token = null;
  state.user = null;
  localStorage.removeItem('token');
  localStorage.removeItem('user');
  updateUI();
  showToast('Вы вышли', 'success');
}

// ===== Link Functions =====

async function addLink() {
  const urlInput = document.getElementById('urlInput');
  const url = urlInput.value.trim();

  if (!url) {
    showToast('Введите URL', 'warning');
    urlInput.focus();
    return;
  }

  try {
    showToast('Добавление...', 'info');
    await api.post('/links', {
      url: url,
      tags: getCurrentTags(),
    });

    urlInput.value = '';
    clearTags();
    showToast('Ссылка добавлена!', 'success');
    await loadLinks();
    await loadStats();
  } catch (e) {}
}

async function loadLinks() {
  try {
    const search = state.filters.search;
    const tags = state.filters.tags;

    let url = '/links';
    const params = new URLSearchParams();

    if (search) params.append('search', search);
    if (tags.length > 0) params.append('tags', tags.join(','));

    if (params.toString()) url += '?' + params.toString();

    console.log('Запрашиваем:', url); // Для отладки
    const data = await api.get(url);
    state.links = data.links || [];
    renderLinks();
  } catch (e) {
    console.error('Ошибка загрузки ссылок:', e);
    state.links = [];
    renderLinks();
  }
}

async function loadStats() {
  try {
    const data = await api.get('/stats');
    document.getElementById('totalLinks').textContent = data.total_links || 0;
    document.getElementById('totalTags').textContent = data.unique_tags || 0;
    document.getElementById('readingTime').textContent = `${data.total_reading_time || 0}м`;
  } catch (e) {
    document.getElementById('totalLinks').textContent = '0';
    document.getElementById('totalTags').textContent = '0';
    document.getElementById('readingTime').textContent = '0м';
  }
}

async function deleteLink(id) {
  if (!confirm('Удалить эту ссылку?')) return;
  try {
    await api.delete(`/links/${id}`);
    showToast('Ссылка удалена', 'success');
    await loadLinks();
    await loadStats();
  } catch (e) {
    // Ошибка уже обработана в api.delete
  }
}

// ===== Export Function =====
async function exportData(format = 'json') {
  try {
    showToast('Подготовка экспорта...', 'info');
    const res = await fetch(`/api/export?format=${format}`, {
      headers: { Authorization: `Bearer ${state.token}` },
    });

    if (!res.ok) throw new Error('Ошибка экспорта');

    const blob = await res.blob();
    const a = document.createElement('a');
    a.href = URL.createObjectURL(blob);
    a.download = `devbrain-export.${format}`;
    a.click();
    showToast('Экспорт готов!', 'success');
  } catch (e) {
    showToast(e.message, 'error');
  }
}

// ===== Stats & Clipboard =====
function showStats() {
  showToast(`Всего ссылок: ${state.links.length}`, 'info');
}

async function pasteFromClipboard() {
  try {
    const text = await navigator.clipboard.readText();
    document.getElementById('urlInput').value = text;
    showToast('Вставлено', 'success');
  } catch (e) {
    showToast('Не удалось вставить', 'error');
  }
}

// ===== UI Functions =====

function renderLinks() {
  const grid = document.getElementById('linksGrid');
  const emptyState = document.getElementById('emptyState');

  if (!grid) return;

  if (state.links.length === 0) {
    grid.innerHTML = '';
    if (emptyState) emptyState.style.display = 'block';
    return;
  }

  if (emptyState) emptyState.style.display = 'none';

  grid.innerHTML = state.links
    .map(
      (link) => `
        <article class="link-card fade-in">
            <div class="link-card-header">
                <div>
                    <h3><a href="${escapeHtml(link.url)}" target="_blank">${escapeHtml(link.title || 'Без названия')}</a></h3>
                    <div class="url">${escapeHtml(link.url)}</div>
                </div>
            </div>
            ${link.description ? `<p class="link-card-description">${escapeHtml(link.description)}</p>` : ''}
            ${
              link.tags && link.tags.length > 0
                ? `
                <div class="link-card-tags">
                    ${link.tags.map((tag) => `<span class="link-card-tag">#${escapeHtml(tag)}</span>`).join('')}
                </div>
            `
                : ''
            }
            <div class="link-card-meta">
                <span>${formatDate(link.created_at)}</span>
                <button class="card-btn delete" onclick="deleteLink(${link.id})">
                    <i class="fas fa-trash"></i>
                </button>
            </div>
        </article>
    `
    )
    .join('');
}

function openAuthModal() {
  const modal = document.getElementById('authModal');
  if (modal) modal.classList.add('active');
}

function closeAuthModal() {
  const modal = document.getElementById('authModal');
  if (modal) modal.classList.remove('active');
}

function switchAuthTab(tab) {
  const loginForm = document.getElementById('loginForm');
  const registerForm = document.getElementById('registerForm');
  const tabs = document.querySelectorAll('.tab-btn');

  if (tab === 'login') {
    if (loginForm) loginForm.style.display = 'block';
    if (registerForm) registerForm.style.display = 'none';
    if (tabs[0]) tabs[0].classList.add('active');
    if (tabs[1]) tabs[1].classList.remove('active');
  } else {
    if (loginForm) loginForm.style.display = 'none';
    if (registerForm) registerForm.style.display = 'block';
    if (tabs[0]) tabs[0].classList.remove('active');
    if (tabs[1]) tabs[1].classList.add('active');
  }
}

function toggleTheme() {
  const html = document.documentElement;
  const current = html.getAttribute('data-theme');
  const next = current === 'light' ? 'dark' : 'light';
  html.setAttribute('data-theme', next);
  localStorage.setItem('theme', next);
  const icon = document.getElementById('themeIcon');
  if (icon) icon.className = next === 'light' ? 'fas fa-moon' : 'fas fa-sun';
}

function updateUI() {
  const authBtn = document.getElementById('authBtn');
  const userMenu = document.getElementById('userMenu');

  if (state.user) {
    if (authBtn) authBtn.style.display = 'none';
    if (userMenu) userMenu.style.display = 'inline-flex';
  } else {
    if (authBtn) authBtn.style.display = 'inline-flex';
    if (userMenu) userMenu.style.display = 'none';
  }
}

function showToast(message, type = 'info') {
  const container = document.getElementById('toastContainer');
  if (!container) return;

  const toast = document.createElement('div');
  toast.className = `toast ${type}`;
  toast.innerHTML = `<span>${escapeHtml(message)}</span>`;
  container.appendChild(toast);
  setTimeout(() => toast.remove(), 3000);
}

// ===== Tags =====
let currentTags = [];

function addTag(tag) {
  tag = tag.trim().toLowerCase();
  if (tag && !currentTags.includes(tag)) {
    currentTags.push(tag);
    renderTags();
  }
}

function removeTag(tag) {
  currentTags = currentTags.filter((t) => t !== tag);
  renderTags();
}

function renderTags() {
  const container = document.getElementById('tagsContainer');
  const input = document.getElementById('tagInput');
  if (!container || !input) return;

  container.querySelectorAll('.tag-chip').forEach((chip) => chip.remove());

  currentTags.forEach((tag) => {
    const chip = document.createElement('div');
    chip.className = 'tag-chip';
    chip.innerHTML = `${tag}<button type="button" onclick="removeTag('${tag}')">&times;</button>`;
    container.insertBefore(chip, input);
  });
}

function getCurrentTags() {
  return [...currentTags];
}

function clearTags() {
  currentTags = [];
  renderTags();
}

// ===== Settings Button =====
function toggleSettings() {
  showToast('Настройки в разработке', 'info');
}

// ===== Utils =====
function escapeHtml(text) {
  if (!text) return '';
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

function formatDate(dateString) {
  if (!dateString) return '';
  const date = new Date(dateString);
  return new Intl.DateTimeFormat('ru-RU', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  }).format(date);
}

// ===== Init =====
document.addEventListener('DOMContentLoaded', () => {
  // Theme
  const savedTheme = localStorage.getItem('theme') || 'light';
  document.documentElement.setAttribute('data-theme', savedTheme);
  const themeIcon = document.getElementById('themeIcon');
  if (themeIcon) themeIcon.className = savedTheme === 'light' ? 'fas fa-moon' : 'fas fa-sun';

  // Auth check
  if (state.token) {
    try {
      state.user = JSON.parse(localStorage.getItem('user') || 'null');
      updateUI();
      loadLinks();
      loadStats();
    } catch (e) {
      state.token = null;
      localStorage.removeItem('token');
    }
  }

  // Forms
  const loginForm = document.getElementById('loginForm');
  if (loginForm) {
    loginForm.addEventListener('submit', (e) => {
      e.preventDefault();
      login(
        document.getElementById('loginEmail').value,
        document.getElementById('loginPassword').value
      );
    });
  }

  const registerForm = document.getElementById('registerForm');
  if (registerForm) {
    registerForm.addEventListener('submit', (e) => {
      e.preventDefault();
      register(
        document.getElementById('regName').value,
        document.getElementById('regEmail').value,
        document.getElementById('regPassword').value
      );
    });
  }

  // Tags
  const tagInput = document.getElementById('tagInput');
  if (tagInput) {
    tagInput.addEventListener('keydown', (e) => {
      if (e.key === 'Enter') {
        e.preventDefault();
        addTag(e.target.value);
        e.target.value = '';
      }
    });
  }

  // Search
  const searchInput = document.getElementById('searchInput');
  if (searchInput) {
    searchInput.addEventListener('input', (e) => {
      state.filters.search = e.target.value;
      loadLinks();
    });
  }
  function toggleSettings() {
    showToast('Настройки в разработке', 'info');
  }
  // Keyboard shortcuts
  document.addEventListener('keydown', (e) => {
    if (e.ctrlKey && e.key === 'k') {
      e.preventDefault();
      if (searchInput) searchInput.focus();
    }
    if (e.ctrlKey && e.key === 'n') {
      e.preventDefault();
      const urlInput = document.getElementById('urlInput');
      if (urlInput) urlInput.focus();
    }
    if (e.key === 'Escape') closeAuthModal();
  });

});
