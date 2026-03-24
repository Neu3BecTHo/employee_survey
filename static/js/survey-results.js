// survey-results.js - Survey results display logic

class SurveyResultsPage {
    constructor() {
        this.surveyId = this.getSurveyIdFromUrl();
        this.init();
    }

    getSurveyIdFromUrl() {
        const pathParts = window.location.pathname.split('/');
        const surveyIndex = pathParts.indexOf('surveys');
        return parseInt(pathParts[surveyIndex + 1]);
    }

    async init() {
        await this.loadResults();
    }

    async loadResults() {
        const loading = document.getElementById('loading');
        const errorMessage = document.getElementById('error-message');
        const resultsContainer = document.getElementById('results-container');

        if (!loading || !errorMessage || !resultsContainer) {
            console.error('Required DOM elements not found');
            return;
        }

        try {
            loading.style.display = 'block';
            errorMessage.style.display = 'none';
            resultsContainer.style.display = 'none';

            const results = await window.app.apiRequest(`/api/surveys/${this.surveyId}/results`);

            // Update survey info with null checks
            const titleElement = document.getElementById('survey-title');
            const descElement = document.getElementById('survey-description');
            const totalElement = document.getElementById('total-responses');
            
            if (titleElement) titleElement.textContent = results.survey.title;
            if (descElement) descElement.textContent = results.survey.description || '';
            if (totalElement) totalElement.textContent = `${results.total_responses} ответов`;

            // Render question results
            this.renderQuestionResults(results.question_results);

            loading.style.display = 'none';
            resultsContainer.style.display = 'block';

        } catch (error) {
            console.error('Error loading results:', error);
            loading.style.display = 'none';
            errorMessage.style.display = 'block';
            if (errorMessage) {
                errorMessage.textContent = 'Ошибка загрузки результатов. Возможно, у вас нет доступа к этим результатам.';
            }
        }
    }

    renderQuestionResults(questionResults) {
        const questionsResults = document.getElementById('questions-results');
        
        if (!questionsResults) {
            console.error('questions-results element not found');
            return;
        }
        
        questionsResults.innerHTML = '';

        if (!questionResults || questionResults.length === 0) {
            questionsResults.innerHTML = '<div class="empty-state"><h2>Результатов не найдено</h2><p>У этого опроса пока нет результатов.</p></div>';
            return;
        }

        questionResults.forEach(result => {
            const questionResultDiv = document.createElement('div');
            questionResultDiv.className = 'question-result';

            let resultsHTML = `<h4>${result.question.text}</h4>`;

            if (result.question.type === 'single_choice') {
                resultsHTML += '<div class="answer-stats">';
                if (result.answers && result.answers.length > 0) {
                    result.answers.forEach(answer => {
                        const percentage = result.question.survey_id ? Math.round((answer.count / this.getTotalResponses(result.answers)) * 100) : 0;
                        resultsHTML += `
                            <div class="answer-stat">
                                <span class="answer-text">${answer.value}</span>
                                <span class="answer-count">${answer.count} (${percentage}%)</span>
                            </div>
                        `;
                    });
                } else {
                    resultsHTML += '<p class="no-answers">Пока нет ответов</p>';
                }
                resultsHTML += '</div>';
            } else if (result.question.type === 'text') {
                resultsHTML += '<div class="answer-stats">';
                if (result.answers && result.answers.length > 0) {
                    result.answers.forEach(answer => {
                        resultsHTML += `
                            <div class="answer-stat">
                                <span class="answer-text">${answer.value}</span>
                                <span class="answer-count">1 ответ</span>
                            </div>
                        `;
                    });
                } else {
                    resultsHTML += '<p class="no-answers">Пока нет ответов</p>';
                }
                resultsHTML += '</div>';
            }

            questionResultDiv.innerHTML = resultsHTML;
            questionsResults.appendChild(questionResultDiv);
        });
    }

    getTotalResponses(answers) {
        return answers.reduce((total, answer) => total + answer.count, 0);
    }
}

// Initialize survey results page when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    new SurveyResultsPage();
});
