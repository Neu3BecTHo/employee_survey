// surveys.js - Main surveys page functionality
class SurveysPage {
    constructor() {
        this.surveys = [];
        this.userResponses = [];
        this.currentUser = null;
        this.init();
    }

    async init() {
        this.setupEventListeners();
        this.loadCurrentUser();
        await this.loadUserResponses();
        await this.loadSurveys();
    }

    setupEventListeners() {
        // Create survey button
        const createBtn = document.getElementById('create-survey-btn');
        if (createBtn) {
            createBtn.addEventListener('click', () => {
                if (window.surveyApp) {
                    window.surveyApp.showCreateSurveyModal();
                }
            });
        }

        // Logout button
        const logoutBtn = document.getElementById('logout-btn');
        if (logoutBtn) {
            logoutBtn.addEventListener('click', () => window.app.logout());
        }
    }

    loadCurrentUser() {
        if (window.app && window.app.currentUser) {
            this.currentUser = window.app.currentUser;
            this.updateUIForUserRole();
        }
    }

    updateUIForUserRole() {
        const createBtn = document.getElementById('create-survey-btn');
        if (createBtn && this.currentUser) {
            if (this.currentUser.role === 'admin') {
                createBtn.classList.remove('hidden');
            } else {
                createBtn.classList.add('hidden');
            }
        }
    }

    async loadUserResponses() {
        try {
            if (window.app && window.app.isLoggedIn()) {
                this.userResponses = await window.app.apiRequest('/surveys/my/data');
            }
        } catch (error) {
        }
    }

    async loadSurveys() {
        const loading = document.getElementById('loading');
        const errorMessage = document.getElementById('error-message');
        const surveysContainer = document.getElementById('surveys-container');

        try {
            loading.classList.remove('hidden');
            errorMessage.classList.add('hidden');
            surveysContainer.classList.add('hidden');

            this.surveys = await window.app.apiRequest('/surveys');
            
            loading.classList.add('hidden');
            surveysContainer.classList.remove('hidden');
            this.renderSurveys();

        } catch (error) {
            loading.classList.add('hidden');
            errorMessage.classList.remove('hidden');
        }
    }

    renderSurveys() {
        const container = document.getElementById('surveys-list');
        container.innerHTML = '';

        if (this.surveys.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <h2>Опросов пока нет</h2>
                    <p>${this.currentUser && this.currentUser.role === 'admin' ? 'Создайте первый опрос для начала работы.' : 'Опросы появятся здесь, когда администраторы их создадут.'}</p>
                    ${this.currentUser && this.currentUser.role === 'admin' ? '<button class="btn btn-primary" onclick="window.surveyApp.showCreateSurveyModal()">➕ Создать опрос</button>' : ''}
                </div>
            `;
            return;
        }

        this.surveys.forEach(survey => {
            const hasResponded = this.userResponses && Array.isArray(this.userResponses) && this.userResponses.some(response => response.survey_id === survey.id);
            const surveyCard = this.createSurveyCard(survey, hasResponded);
            container.appendChild(surveyCard);
        });
    }

    createSurveyCard(survey, hasResponded) {
        const card = document.createElement('div');
        card.className = 'survey-card';
        
        const statusBadge = this.getStatusBadge(survey.status);
        const actionButton = this.getActionButton(survey, hasResponded);
        
        card.innerHTML = `
            <div class="survey-card-content">
                <div class="survey-header">
                    <h3>${survey.title}</h3>
                    <div class="survey-meta">
                        ${statusBadge}
                        <span class="survey-date">${new Date(survey.created_at).toLocaleDateString('ru-RU')}</span>
                    </div>
                </div>
                <div class="survey-content">
                    <p>${survey.description || 'Без описания'}</p>
                </div>
                <div class="survey-actions">
                    ${actionButton}
                    ${this.currentUser && this.currentUser.role === 'admin' ? `
                        <a href="/admin/surveys/${survey.id}" class="btn btn-secondary btn-sm">
                            ✏️ Редактировать
                        </a>
                        <a href="/surveys/${survey.id}/results" class="btn btn-secondary btn-sm">
                            📊 Результаты
                        </a>
                    ` : ''}
                </div>
            </div>
        `;
        
        return card;
    }

    getStatusBadge(status) {
        const badges = {
            'draft': '<span class="status-badge draft">📝 Черновик</span>',
            'open': '<span class="status-badge open">📢 Открыт</span>',
            'closed': '<span class="status-badge closed">🔒 Закрыт</span>'
        };
        return badges[status] || badges['draft'];
    }

    getActionButton(survey, hasResponded) {
        if (hasResponded) {
            const userResponse = this.userResponses && Array.isArray(this.userResponses) && this.userResponses.find(r => r.survey_id === survey.id);
            if (userResponse) {
                return `
                    <a href="/surveys/responses/${userResponse.id}" class="btn btn-secondary btn-sm">
                        📝 Мои ответы
                    </a>
                `;
            }
        }

        if (survey.status === 'open') {
            return `
                <a href="/surveys/${survey.id}/take" class="btn btn-primary btn-sm">
                    📝 Пройти опрос
                </a>
            `;
        }

        if (survey.status === 'closed') {
            return `
                <a href="/surveys/${survey.id}/results" class="btn btn-secondary btn-sm">
                    📊 Результаты
                </a>
            `;
        }

        return '<span class="btn btn-secondary btn-sm" disabled>📝 Недоступен</span>';
    }
}

// Initialize page when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.surveysPage = new SurveysPage();
});
