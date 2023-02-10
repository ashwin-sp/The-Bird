<script setup>
import { storeToRefs } from 'pinia';
import { useAuthStore } from '@/stores';
import { fetchWrapper } from '@/helpers';
import Post from '../components/Post.vue';

const authStore = useAuthStore();
const { user } = storeToRefs(authStore);

const baseUrl = `${import.meta.env.VITE_API_URL}`;
const userInfo = JSON.parse(localStorage.getItem('user'))
let feed = await fetchWrapper.post(`${baseUrl}/viewpersonalfeeds`, {"username": userInfo.username}).then( (data) => {
    console.log("Inside", data);
    return data.json();});

console.log("ouside", feed);

</script>

<template>
    <div v-if="user">
        <div class="album py-5 bg-light">
            <div class="container">
                <div class="row row-cols-1">
                    <h3>Your Personal Feed:</h3>
                    <div class="col" v-for="item in feed" :key="item.postid">
                        <Post :Timestamp="item.timestamp" :Username="item.username"
                            :Message="item.message" />
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>
