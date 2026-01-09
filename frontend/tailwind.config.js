/** @type {import('tailwindcss').Config} */
export default {
    content: [
        "./index.html",
        "./src/**/*.{js,ts,jsx,tsx}",
    ],
    theme: {
        extend: {
            colors: {
                parchment: {
                    50: '#F9F4EF',
                    100: '#F2E8DC', // Main Background
                    200: '#E6DCC9', // Secondary Background
                    300: '#D9CDB1',
                },
                ink: {
                    50: '#F5F5F4',
                    800: '#3E2723', // Sepia
                    900: '#2D2A26', // Charcoal
                },
                wax: {
                    red: '#8B0000',
                },
                bronze: {
                    DEFAULT: '#CD853F',
                }
            },
            fontFamily: {
                serif: ['Merriweather', 'serif'],
                sans: ['Inter', 'sans-serif'],
            },
        },
    },
    plugins: [],
}
