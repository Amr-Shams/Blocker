// Particle animation
function createParticles() {
    const particles = document.getElementById('particles');
    const particleCount = 50;

    for (let i = 0; i < particleCount; i++) {
        const particle = document.createElement('div');
        particle.className = 'particle';

        const size = Math.random() * 3 + 1;
        const initialX = Math.random() * window.innerWidth;
        const initialY = Math.random() * window.innerHeight;
        const duration = Math.random() * 20 + 10;
        const delay = Math.random() * -20;

        particle.style.width = `${size}px`;
        particle.style.height = `${size}px`;
        particle.style.left = `${initialX}px`;
        particle.style.top = `${initialY}px`;

        particle.style.animation = `floatParticle ${duration}s linear ${delay}s infinite`;
        particles.appendChild(particle);
    }
}

// Wallet interaction
const pocketWallet = document.getElementById('pocketWallet');
const balanceDisplay = document.getElementById('balanceDisplay');
const loginForm = document.getElementById('loginForm');
const loginButton = document.getElementById('loginButton');
let isOpen = false;

pocketWallet.addEventListener('click', () => {
    if (!isOpen) {
        pocketWallet.classList.add('open');
        isOpen = true;

        setTimeout(() => {
            balanceDisplay.classList.add('hidden');
            loginForm.classList.add('active');
        }, 1000);
    }
});

loginButton.addEventListener('click', () => {
    loginButton.style.transform = 'scale(0.95)';
    setTimeout(() => {
        loginButton.style.transform = 'scale(1)';
        window.location.href = '/dashboard'; // Update this URL as needed
    }, 200);
});

// Floating animation
const walletContainer = document.getElementById('walletContainer');
let floating = false;
let floatY = 0;
let floatSpeed = 0.02;

function floatAnimation() {
    if (floating) {
        floatY += floatSpeed;
        if (floatY > 1 || floatY < 0) floatSpeed *= -1;
        walletContainer.style.transform = `translateY(${Math.sin(floatY * Math.PI) * 10}px)`;
    }
    requestAnimationFrame(floatAnimation);
}

walletContainer.addEventListener('mouseenter', () => {
    floating = true;
});

walletContainer.addEventListener('mouseleave', () => {
    floating = false;
    walletContainer.style.transform = '';
});

// Initialize
createParticles();
floatAnimation();

// Add floating animation style
document.head.insertAdjacentHTML('beforeend', `
    <style>
        @keyframes floatParticle {
            0% {
                transform: translateY(100vh) scale(1);
                opacity: 0;
            }
            10% {
                opacity: 1;
            }
            90% {
                opacity: 1;
            }
            100% {
                transform: translateY(-100px) scale(0);
                opacity: 0;
            }
        }
    </style>
`);

