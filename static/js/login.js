// Login Page with Enhanced UI
class LoginPage {
    constructor() {
        this.users = [];
        this.init();
    }

    init() {
        this.loadUsers();
    }

    async loadUsers() {
        const loading = document.getElementById('loading');
        const errorMessage = document.getElementById('error-message');
        const usersContainer = document.getElementById('users-container');

        try {
            loading.style.display = 'block';
            errorMessage.style.display = 'none';
            usersContainer.style.display = 'none';

            const users = await window.app.apiRequest('/users');
            this.users = users;

            loading.style.display = 'none';
            usersContainer.style.display = 'grid';
            this.renderUsers();

        } catch (error) {
            console.error('Error loading users:', error);
            loading.style.display = 'none';
            errorMessage.style.display = 'block';
        }
    }

    renderUsers() {
        const container = document.getElementById('users-container');
        
        container.innerHTML = this.users.map((user, index) => `
            <div class="user-card" onclick="loginPage.selectUser(${user.id}, '${user.name}', '${user.role}')" style="animation-delay: ${index * 0.1}s">
                <h4>${user.name}</h4>
                <div class="user-role">${user.role === 'admin' ? '👑 Администратор' : '👤 Сотрудник'}</div>
            </div>
        `).join('');
    }

    selectUser(userId, userName, userRole) {
        // Store user info with consistent keys
        localStorage.setItem('user_id', userId);
        localStorage.setItem('user_name', userName);
        localStorage.setItem('user_role', userRole);

        // Show success message
        this.showSuccessMessage(`Вы вошли как ${userName}`);

        // Redirect to main page after a short delay
        setTimeout(() => {
            window.location.href = '/';
        }, 1000);
    }

    showSuccessMessage(message) {
        // Remove existing messages
        const existingMessage = document.querySelector('.message-toast');
        if (existingMessage) {
            existingMessage.remove();
        }

        // Create new message
        const messageDiv = document.createElement('div');
        messageDiv.className = 'message-toast message-success';
        messageDiv.innerHTML = `
            <span class="message-icon">✅</span>
            <span class="message-text">${message}</span>
        `;

        // Add styles
        messageDiv.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            background: var(--success);
            color: white;
            padding: 12px 20px;
            border-radius: 8px;
            box-shadow: var(--shadow-lg);
            z-index: 3000;
            display: flex;
            align-items: center;
            gap: 8px;
            animation: slideInRight 0.3s ease-out;
        `;

        document.body.appendChild(messageDiv);
    }
}

// Add animation styles
const style = document.createElement('style');
style.textContent = `
    @keyframes slideInRight {
        from {
            transform: translateX(100%);
            opacity: 0;
        }
        to {
            transform: translateX(0);
            opacity: 1;
        }
    }

    .user-card {
        animation: fadeInUp 0.3s ease-out;
        animation-fill-mode: both;
    }

    @keyframes fadeInUp {
        from {
            transform: translateY(20px);
            opacity: 0;
        }
        to {
            transform: translateY(0);
            opacity: 1;
        }
    }
`;
document.head.appendChild(style);

// Initialize page when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.loginPage = new LoginPage();
});
