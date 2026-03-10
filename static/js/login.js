// login.js - Login page logic

class LoginPage {
    constructor() {
        this.init();
    }

    async init() {
        await this.loadUsers();
    }

    async loadUsers() {
        const loading = document.getElementById('loading');
        const errorMessage = document.getElementById('error-message');
        const userList = document.getElementById('user-list');
        const usersContainer = document.getElementById('users-container');

        try {
            loading.style.display = 'block';
            errorMessage.style.display = 'none';
            userList.style.display = 'none';

            const response = await fetch('/users');
            if (!response.ok) {
                throw new Error('Failed to load users');
            }

            const users = await response.json();

            // Render users
            usersContainer.innerHTML = '';
            users.forEach(user => {
                const userCard = document.createElement('div');
                userCard.className = 'user-card';
                userCard.onclick = () => this.loginAsUser(user);

                userCard.innerHTML = `
                    <h3>${user.name}</h3>
                    <p>Роль: ${this.getRoleText(user.role)}</p>
                `;

                usersContainer.appendChild(userCard);
            });

            loading.style.display = 'none';
            userList.style.display = 'block';

        } catch (error) {
            console.error('Error loading users:', error);
            loading.style.display = 'none';
            errorMessage.style.display = 'block';
        }
    }

    getRoleText(role) {
        const roleMap = {
            'admin': 'Администратор',
            'employee': 'Сотрудник'
        };
        return roleMap[role] || role;
    }

    loginAsUser(user) {
        // Store user info in localStorage
        localStorage.setItem('user_id', user.id.toString());
        localStorage.setItem('user_name', user.name);
        localStorage.setItem('user_role', user.role);

        // Redirect to main page
        window.location.href = '/';
    }
}

// Initialize login page when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    new LoginPage();
});
