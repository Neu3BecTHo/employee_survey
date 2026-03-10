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

        try {
            loading.style.display = 'block';
            errorMessage.style.display = 'none';
            resultsContainer.style.display = 'none';

            const results = await window.app.apiRequest(`/surveys/${this.surveyId}/results`);

            // Update survey info
            document.getElementById('survey-title').textContent = results.survey.title;
            document.getElementById('survey-description').textContent = results.survey.description || '';
            document.getElementById('total-responses').textContent = results.total_responses;
            document.getElementById('survey-status').textContent = window.app.getStatusText(results.survey.status);

            // Render question results
            this.renderQuestionResults(results.question_results);

            loading.style.display = 'none';
            resultsContainer.style.display = 'block';

        } catch (error) {
            console.error('Error loading results:', error);
            loading.style.display = 'none';
            errorMessage.style.display = 'block';
            errorMessage.textContent = 'Ошибка загрузки результатов. Возможно, у вас нет доступа к этим результатам.';
        }
    }

    renderQuestionResults(questionResults) {
        const questionsResults = document.getElementById('questions-results');
        questionsResults.innerHTML = '';

        questionResults.forEach(result => {
            const questionResultDiv = document.createElement('div');
            questionResultDiv.className = 'question-result';

            let resultsHTML = `<h4>${result.question.text}</h4>`;

            if (result.question.type === 'single_choice') {
                resultsHTML += '<div class="answer-stats">';
                result.answers.forEach(answer => {
                    const percentage = result.question.survey_id ? Math.round((answer.count / this.getTotalResponses(result.answers)) * 100) : 0;
                    resultsHTML += `
                        <div class="answer-stat">
                            <span class="answer-text">${answer.value}</span>
                            <span class="answer-count">${answer.count} (${percentage}%)</span>
                        </div>
                    `;
                });
                resultsHTML += '</div>';
            } else if (result.question.type === 'text') {
                resultsHTML += '<div class="answer-stats">';
                result.answers.forEach(answer => {
                    resultsHTML += `
                        <div class="answer-stat">
                            <span class="answer-text">${answer.value}</span>
                            <span class="answer-count">1 ответ</span>
                        </div>
                    `;
                });
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
