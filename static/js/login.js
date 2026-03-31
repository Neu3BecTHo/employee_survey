// Login Page with Enhanced UI
class LoginPage {
    constructor() {
        this.users = [];
        this.init();
    }

    init() {
        // Check if user is already logged in - redirect to home
        if (localStorage.getItem('user_id')) {
            window.location.href = '/';
            return;
        }
        this.loadUsers();
    }

    async loadUsers() {
        const loading = document.getElementById('loading');
        const errorMessage = document.getElementById('error-message');
        const usersContainer = document.getElementById('users-container');

        try {
            loading.classList.remove('hidden');
            errorMessage.classList.add('hidden');
            usersContainer.classList.add('hidden');

            const users = await window.app.apiRequest('/users');
            this.users = users;

            loading.classList.add('hidden');
            usersContainer.classList.remove('hidden');
            this.renderUsers();

        } catch (error) {
            loading.classList.add('hidden');
            errorMessage.classList.remove('hidden');
        }
    }

    renderUsers() {
        const container = document.getElementById('users-container');
        
        container.innerHTML = this.users.map((user, index) => `
            <div class="user-card" onclick="loginPage.selectUser(${user.id}, '${user.name}', '${user.role}')" style="animation-delay: ${index * 0.1}s">
                <div class="user-name">${user.name}</div>
                <div class="user-role">${user.role === 'admin' ? 'Администратор' : 'Сотрудник'}</div>
            </div>
        `).join('');
    }

    selectUser(userId, userName, userRole) {
        // Store user info with consistent keys
        localStorage.setItem('user_id', userId);
        localStorage.setItem('user_name', userName);
        localStorage.setItem('user_role', userRole);

        // Show success message
        window.app.showMessage(`Вы вошли как ${userName}`, 'success', 'Вход выполнен');

        // Redirect to main page after a short delay
        setTimeout(() => {
            window.location.href = '/';
        }, 1000);
    }
}

// Initialize login page when DOM is ready
// Initialize page when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.loginPage = new LoginPage();
});
