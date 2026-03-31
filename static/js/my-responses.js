// My Responses Page
class MyResponsesPage {
    constructor() {
        this.responses = [];
        this.init();
    }

    async init() {
        document.getElementById('logout-btn').addEventListener('click', () => {
            window.app.logout();
        });

        await this.loadMyResponses();
    }

    async loadMyResponses() {
        const loading = document.getElementById('loading');
        const errorMessage = document.getElementById('error-message');
        const responsesContainer = document.getElementById('responses-container');
        const responsesList = document.getElementById('responses-list');

        try {
            loading.classList.remove('hidden');
            errorMessage.classList.add('hidden');
            responsesContainer.classList.add('hidden');

            const responses = await window.app.apiRequest('/surveys/my/data');
            this.responses = responses;

            loading.classList.add('hidden');
            responsesContainer.classList.remove('hidden');

            if (!responses || responses.length === 0) {
                responsesList.innerHTML = '<div class="empty-state"><h2>Ответов пока нет</h2><p>Вы еще не прошли ни одного опроса.</p></div>';
                return;
            }

            this.renderResponses();

        } catch (error) {
            loading.classList.add('hidden');
            errorMessage.classList.remove('hidden');
            errorMessage.textContent = 'Ошибка загрузки ответов.';
        }
    }

    renderResponses() {
        const responsesList = document.getElementById('responses-list');
        responsesList.innerHTML = '';
        responsesList.className = 'surveys-grid'; // Use same grid as surveys

        this.responses.forEach(response => {
            const responseCard = document.createElement('div');
            responseCard.className = 'response-card';
            
            responseCard.innerHTML = `
                <div class="response-header">
                    <h3>${response.survey_title}</h3>
                    <div class="response-date">
                        ${new Date(response.submitted_at).toLocaleDateString('ru-RU')}
                    </div>
                </div>
                <div class="response-content">
                    <p>${response.survey_description || 'Без описания'}</p>
                    <div class="response-actions">
                        <a href="/surveys/responses/${response.id}" class="btn btn-primary btn-sm">
                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
                                <polyline points="14 2 14 8 20 8"></polyline>
                                <line x1="16" y1="13" x2="8" y2="13"></line>
                            </svg>
                            Мои ответы
                        </a>
                        <a href="/surveys/${response.survey_id}/results" class="btn btn-secondary btn-sm">
                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                <path d="M3 3v18h18"></path>
                                <rect x="7" y="10" width="4" height="9"></rect>
                                <rect x="15" y="6" width="4" height="13"></rect>
                            </svg>
                            Результаты
                        </a>
                    </div>
                </div>
            `;
            responsesList.appendChild(responseCard);
        });
    }
}

// Initialize page when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    new MyResponsesPage();
});
