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
            
            console.log('Creating survey with data:', data);
            console.log('Current user:', this.currentUser);
            
            const response = await this.apiRequest('/api/surveys', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(data)
            });
            
            console.log('Survey created response:', response);

            // Reload surveys list
            if (window.surveysPage) {
                window.surveysPage.loadSurveys();
            }

            this.showMessage('Опрос успешно создан!', 'success');
        } catch (error) {
            console.error('Error creating survey:', error);
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

    showMessage(message, type = 'info') {
        // Remove existing messages
        const existingMessages = document.querySelectorAll('.message');
        existingMessages.forEach(msg => msg.remove());

        const messageDiv = document.createElement('div');
        messageDiv.className = `message message-${type}`;
        messageDiv.textContent = message;

        // Style the message
        messageDiv.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            padding: 1rem;
            border-radius: 6px;
            color: white;
            font-weight: 500;
            z-index: 1000;
            max-width: 300px;
        `;

        if (type === 'success') {
            messageDiv.style.backgroundColor = '#28a745';
        } else if (type === 'error') {
            messageDiv.style.backgroundColor = '#dc3545';
        } else {
            messageDiv.style.backgroundColor = '#007bff';
        }

        document.body.appendChild(messageDiv);

        // Auto remove after 3 seconds
        setTimeout(() => {
            if (messageDiv.parentNode) {
                messageDiv.remove();
            }
        }, 3000);
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
