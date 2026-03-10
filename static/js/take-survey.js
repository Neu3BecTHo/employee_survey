// take-survey.js - Survey taking logic

class TakeSurveyPage {
    constructor() {
        this.surveyId = this.getSurveyIdFromUrl();
        this.survey = null;
        this.init();
    }

    getSurveyIdFromUrl() {
        const pathParts = window.location.pathname.split('/');
        return parseInt(pathParts[pathParts.length - 2]);
    }

    async init() {
        await this.loadSurvey();
    }

    async loadSurvey() {
        const loading = document.getElementById('loading');
        const errorMessage = document.getElementById('error-message');
        const surveyContainer = document.getElementById('survey-container');

        try {
            loading.style.display = 'block';
            errorMessage.style.display = 'none';
            surveyContainer.style.display = 'none';

            const surveyData = await window.app.apiRequest(`/surveys/${this.surveyId}`);
            this.survey = surveyData;

            // Update title and description
            document.getElementById('survey-title').textContent = this.survey.title;
            document.getElementById('survey-description').textContent = this.survey.description || '';

            // Render questions
            this.renderQuestions();

            loading.style.display = 'none';
            surveyContainer.style.display = 'block';

            // Setup form submission
            this.setupFormSubmission();

        } catch (error) {
            console.error('Error loading survey:', error);
            loading.style.display = 'none';
            errorMessage.style.display = 'block';
            errorMessage.textContent = 'Ошибка загрузки опроса. Возможно, опрос не найден или закрыт.';
        }
    }

    renderQuestions() {
        const questionsContainer = document.getElementById('questions-container');
        questionsContainer.innerHTML = '';

        this.survey.questions.forEach((question, index) => {
            const questionCard = document.createElement('div');
            questionCard.className = `question-card ${question.is_required ? 'required' : ''}`;

            let questionHTML = `
                <h4>${question.text}${question.is_required ? ' <span class="required-mark">*</span>' : ''}</h4>
            `;

            if (question.type === 'single_choice') {
                questionHTML += '<div class="radio-group">';
                question.options.forEach(option => {
                    questionHTML += `
                        <label class="radio-option">
                            <input type="radio" name="question_${question.id}" value="${option}" required="${question.is_required}">
                            ${option}
                        </label>
                    `;
                });
                questionHTML += '</div>';
            } else if (question.type === 'text') {
                questionHTML += `
                    <div class="form-group">
                        <textarea name="question_${question.id}" rows="3" ${question.is_required ? 'required' : ''} placeholder="Введите ваш ответ..."></textarea>
                    </div>
                `;
            }

            questionCard.innerHTML = questionHTML;
            questionsContainer.appendChild(questionCard);
        });
    }

    setupFormSubmission() {
        const form = document.getElementById('survey-form');
        const submitBtn = document.getElementById('submit-btn');
        const successMessage = document.getElementById('success-message');

        form.addEventListener('submit', async (e) => {
            e.preventDefault();

            // Collect answers
            const answers = [];
            const formData = new FormData(form);

            for (const [key, value] of formData.entries()) {
                if (key.startsWith('question_') && value.trim() !== '') {
                    const questionId = parseInt(key.replace('question_', ''));
                    answers.push({
                        question_id: questionId,
                        value: value.trim()
                    });
                }
            }

            // Validate
            if (!this.validateAnswers(answers)) {
                window.app.showMessage('Пожалуйста, заполните все обязательные поля', 'error');
                return;
            }

            try {
                submitBtn.disabled = true;
                submitBtn.textContent = 'Отправка...';

                await window.app.apiRequest(`/surveys/${this.surveyId}/responses`, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({ answers })
                });

                form.style.display = 'none';
                successMessage.style.display = 'block';

                // Redirect to home after 3 seconds
                setTimeout(() => {
                    window.location.href = '/';
                }, 3000);

            } catch (error) {
                console.error('Error submitting survey:', error);
                window.app.showMessage('Ошибка при отправке ответов. Попробуйте еще раз.', 'error');
                submitBtn.disabled = false;
                submitBtn.textContent = 'Отправить ответы';
            }
        });
    }

    validateAnswers(answers) {
        const questionMap = {};
        this.survey.questions.forEach(q => {
            questionMap[q.id] = q;
        });

        // Check required questions are answered
        for (const question of this.survey.questions) {
            if (question.is_required) {
                const answer = answers.find(a => a.question_id === question.id);
                if (!answer || !answer.value) {
                    return false;
                }
            }
        }

        return true;
    }
}

// Initialize take survey page when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    new TakeSurveyPage();
});
