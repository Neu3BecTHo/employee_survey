// app.js - Core application logic

class SurveyApp {
    constructor() {
        this.currentUser = null;
        this.init();
    }

    init() {
        this.loadCurrentUser();
        this.setupEventListeners();
    }

    loadCurrentUser() {
        const userId = localStorage.getItem('user_id');
        const userName = localStorage.getItem('user_name');
        const userRole = localStorage.getItem('user_role');

        if (userId && userName && userRole) {
            this.currentUser = {
                id: parseInt(userId),
                name: userName,
                role: userRole
            };
            this.updateUIForUser();
        } else {
            // No user logged in, redirect to login
            if (!window.location.pathname.includes('/login')) {
                window.location.href = '/login';
            }
        }
    }

    isLoggedIn() {
        return this.currentUser !== null;
    }

    updateUIForUser() {
        const userElement = document.getElementById('current-user');
        const roleElement = document.getElementById('user-role');
        const adminControls = document.getElementById('admin-controls');

        if (userElement) {
            userElement.textContent = this.currentUser.name;
        }

        if (roleElement) {
            roleElement.textContent = this.getRoleText(this.currentUser.role);
            roleElement.className = `role-text`;
        }

        if (adminControls && this.currentUser.role === 'admin') {
            adminControls.style.display = 'block';
        }
    }

    getRoleText(role) {
        const roleTexts = {
            'admin': 'Администратор',
            'employee': 'Сотрудник'
        };
        return roleTexts[role] || role;
    }

    setupEventListeners() {
        // Logout button
        const logoutBtn = document.getElementById('logout-btn');
        if (logoutBtn) {
            logoutBtn.addEventListener('click', () => this.logout());
        }

        // Create survey button
        const createSurveyBtn = document.getElementById('create-survey-btn');
        if (createSurveyBtn) {
            createSurveyBtn.addEventListener('click', () => this.showCreateSurveyModal());
        }
    }

    logout() {
        localStorage.removeItem('user_id');
        localStorage.removeItem('user_name');
        localStorage.removeItem('user_role');
        this.currentUser = null;
        window.location.href = '/login';
    }

    showCreateSurveyModal() {
        const modal = document.getElementById('create-survey-modal');
        const form = document.getElementById('create-survey-form');

        modal.style.display = 'flex';

        // Close modal handlers
        const closeBtn = modal.querySelector('.modal-close');
        const cancelBtn = modal.querySelector('.modal-cancel');

        const closeModal = () => {
            modal.style.display = 'none';
            form.reset();
        };

        closeBtn.addEventListener('click', closeModal);
        cancelBtn.addEventListener('click', closeModal);

        modal.addEventListener('click', (e) => {
            if (e.target === modal) {
                closeModal();
            }
        });

        // Form submission
        form.addEventListener('submit', async (e) => {
            e.preventDefault();
            await this.createSurvey(new FormData(form));
            closeModal();
        });
    }

    hideCreateSurveyModal() {
        const modal = document.getElementById('create-survey-modal');
        const form = document.getElementById('create-survey-form');
        
        if (modal) {
            modal.style.display = 'none';
        }
        if (form) {
            form.reset();
        }
    }

    async createSurvey(formData) {
        try {
            const data = {
                title: formData.get('title'),
                description: formData.get('description'),
                status: 'draft'  // Add default status to satisfy database constraint
            };
            
            
            const response = await this.apiRequest('/surveys', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(data)
            });
            

            // Reload surveys list
            if (window.surveysPage) {
                window.surveysPage.loadSurveys();
            }

            this.showMessage('Опрос успешно создан!', 'success');
        } catch (error) {
            this.showMessage('Ошибка при создании опроса', 'error');
        }
    }

    async apiRequest(url, options = {}) {
        const headers = {
            'X-User-Id': this.currentUser ? this.currentUser.id.toString() : '',
            ...options.headers
        };

        const response = await fetch(url, {
            ...options,
            headers
        });

        if (!response.ok) {
            const error = await response.json().catch(() => ({ error: 'Network error' }));
            throw new Error(error.error || `HTTP ${response.status}`);
        }

        return response.json();
    }

    showMessage(message, type = 'info', title = '') {
        // Remove existing toasts
        const existingToasts = document.querySelectorAll('.toast');
        existingToasts.forEach(toast => {
            toast.classList.add('hiding');
            setTimeout(() => toast.remove(), 300);
        });

        const toast = document.createElement('div');
        toast.className = `toast toast-${type}`;
        
        // Add icon based on type
        const icons = {
            success: '<svg class="toast-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path><polyline points="22 4 12 14.01 9 11.01"></polyline></svg>',
            error: '<svg class="toast-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="15" y1="9" x2="9" y2="15"></line><line x1="9" y1="9" x2="15" y2="15"></line></svg>',
            warning: '<svg class="toast-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path><line x1="12" y1="9" x2="12" y2="13"></line><line x1="12" y1="17" x2="12.01" y2="17"></line></svg>',
            info: '<svg class="toast-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="16" x2="12" y2="12"></line><line x1="12" y1="8" x2="12.01" y2="8"></line></svg>'
        };
        
        const titles = {
            success: 'Успех',
            error: 'Ошибка',
            warning: 'Внимание',
            info: 'Информация'
        };
        
        const displayTitle = title || titles[type] || titles.info;
        
        toast.innerHTML = `
            ${icons[type] || icons.info}
            <div class="toast-content">
                <div class="toast-title">${displayTitle}</div>
                <div class="toast-message">${message}</div>
            </div>
            <button class="toast-close" onclick="this.parentElement.classList.add('hiding'); setTimeout(() => this.parentElement.remove(), 300)">×</button>
            <div class="toast-progress"></div>
        `;

        document.body.appendChild(toast);

        // Auto-remove after 4 seconds
        setTimeout(() => {
            if (toast.parentNode) {
                toast.classList.add('hiding');
                setTimeout(() => toast.remove(), 300);
            }
        }, 4000);
    }

    formatDate(dateString) {
        const date = new Date(dateString);
        return date.toLocaleDateString('ru-RU', {
            year: 'numeric',
            month: 'long',
            day: 'numeric'
        });
    }

    getStatusText(status) {
        const statusMap = {
            'draft': 'Черновик',
            'open': 'Открыт',
            'closed': 'Закрыт'
        };
        return statusMap[status] || status;
    }

    getStatusClass(status) {
        return status.toLowerCase();
    }
}

// Initialize app when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.app = new SurveyApp();
});
