// surveys.js - Surveys list and management logic

class SurveysManager {
    constructor() {
        this.init();
    }

    init() {
        this.loadSurveys();
    }

    async loadSurveys() {
        const loading = document.getElementById('loading');
        const errorMessage = document.getElementById('error-message');
        const emptyState = document.getElementById('empty-state');
        const surveysList = document.getElementById('surveys-list');

        try {
            loading.style.display = 'block';
            errorMessage.style.display = 'none';
            emptyState.style.display = 'none';
            surveysList.style.display = 'none';

            const surveys = await window.app.apiRequest('/surveys');

            loading.style.display = 'none';

            if (surveys.length === 0) {
                emptyState.style.display = 'block';
            } else {
                this.renderSurveys(surveys);
                surveysList.style.display = 'block';
            }

        } catch (error) {
            console.error('Error loading surveys:', error);
            loading.style.display = 'none';
            errorMessage.style.display = 'block';
            errorMessage.textContent = 'Ошибка загрузки опросов. Попробуйте обновить страницу.';
        }
    }

    renderSurveys(surveys) {
        const surveysList = document.getElementById('surveys-list');
        surveysList.innerHTML = '';

        surveys.forEach(survey => {
            const surveyCard = document.createElement('div');
            surveyCard.className = `survey-card ${window.app.getStatusClass(survey.status)}`;

            const actions = this.getSurveyActions(survey);

            surveyCard.innerHTML = `
                <h3>${survey.title}</h3>
                <p>${survey.description || 'Без описания'}</p>
                <div class="survey-meta">
                    <span>Статус: ${window.app.getStatusText(survey.status)}</span>
                    <span>Создан: ${window.app.formatDate(survey.created_at)}</span>
                </div>
                <div class="survey-actions">
                    ${actions}
                </div>
            `;

            surveysList.appendChild(surveyCard);
        });
    }

    getSurveyActions(survey) {
        const userRole = window.app.currentUser.role;
        let actions = '';

        if (userRole === 'admin') {
            if (survey.status === 'draft') {
                actions += `<button class="btn btn-success" onclick="surveysManager.openSurvey(${survey.id})">Открыть</button>`;
                actions += `<button class="btn btn-primary" onclick="window.location.href='/admin/surveys/${survey.id}'">Редактировать</button>`;
            } else if (survey.status === 'open') {
                actions += `<button class="btn btn-danger" onclick="surveysManager.closeSurvey(${survey.id})">Закрыть</button>`;
                actions += `<button class="btn btn-secondary" onclick="window.location.href='/admin/surveys/${survey.id}/results'">Результаты</button>`;
            } else {
                actions += `<button class="btn btn-secondary" onclick="window.location.href='/admin/surveys/${survey.id}/results'">Результаты</button>`;
            }
        } else {
            // Employee actions
            if (survey.status === 'open') {
                actions += `<button class="btn btn-primary" onclick="window.location.href='/surveys/${survey.id}/take'">Пройти опрос</button>`;
            } else if (survey.status === 'closed') {
                actions += `<button class="btn btn-secondary" onclick="window.location.href='/surveys/${survey.id}/results'">Посмотреть ответы</button>`;
            }
        }

        return actions;
    }

    async openSurvey(surveyId) {
        try {
            await window.app.apiRequest(`/surveys/${surveyId}/open`, { method: 'POST' });
            this.loadSurveys();
            window.app.showMessage('Опрос открыт!', 'success');
        } catch (error) {
            console.error('Error opening survey:', error);
            window.app.showMessage('Ошибка при открытии опроса', 'error');
        }
    }

    async closeSurvey(surveyId) {
        try {
            await window.app.apiRequest(`/surveys/${surveyId}/close`, { method: 'POST' });
            this.loadSurveys();
            window.app.showMessage('Опрос закрыт!', 'success');
        } catch (error) {
            console.error('Error closing survey:', error);
            window.app.showMessage('Ошибка при закрытии опроса', 'error');
        }
    }
}

// Global instance for onclick handlers
let surveysManager;

// Initialize surveys manager when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    surveysManager = new SurveysManager();
});

// Export for use in other scripts
window.loadSurveys = () => surveysManager.loadSurveys();
