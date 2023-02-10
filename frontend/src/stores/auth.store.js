import { defineStore } from 'pinia';

import { fetchWrapper } from '@/helpers';
import { router } from '@/router';
import { useAlertStore } from '@/stores';

const baseUrl = `${import.meta.env.VITE_API_URL}`;

export const useAuthStore = defineStore({
    id: 'auth',
    state: () => ({
        user: JSON.parse(localStorage.getItem('user')),
        returnUrl: null
    }),
    actions: {
        async login(username, password) {
            try {
                let user = await fetchWrapper.post(`${baseUrl}/login`, { username, password });    
                user.token= `accesstoken=${user.accesstoken};refreshtoken=${user.refreshtoken}`
                user.username = username

                console.log(user)
                this.user = user;
                localStorage.setItem('user', JSON.stringify(user));

                router.push(this.returnUrl || '/');
            } catch (error) {
                const alertStore = useAlertStore();
                alertStore.error(error);                
            }
        },
        async register(username, password) {
            try {
                let user = await fetchWrapper.post(`${baseUrl}/signup`, { username, password });    
                user.token= `accesstoken=${user.accesstoken};refreshtoken=${user.refreshtoken}`
                // user.username = username
                console.log(user)
                // this.user = user;
                // localStorage.setItem('user', JSON.stringify(user));
                // router.push('/');
            } catch (error) {
                const alertStore = useAlertStore();
                alertStore.error(error);                
            }
        },
        logout() {
            this.user = null;
            localStorage.removeItem('user');
            router.push('/account/login');
        }
    }
});
