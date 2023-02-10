<script setup>
import { Form, Field } from 'vee-validate';
import * as Yup from 'yup';

import { useAlertStore } from '@/stores';
import { fetchWrapper } from '@/helpers';

const alertStore = useAlertStore();

let title = 'UnFollow User';
const baseUrl = `${import.meta.env.VITE_API_URL}`;

const schema = Yup.object().shape({
    username: Yup.string()
        .required('Username is required'),
});

const local_user = JSON.parse(localStorage.getItem('user')).username;

async function onSubmit(values) {
    let data = {
        "username": values.username,
        "follower": local_user,
        "isfollowing": false
    }
    let message;
    let response = await fetchWrapper.post(`${baseUrl}/updatefollow`, data);
    if (response.status == 200) {
        message = 'User UnFollowed';
        alertStore.success(message);
    }
    else {
        message = "Error";
        alertStore.error(message);
    }
}
</script>

<template>
    <h1>{{ title }}</h1>
    <Form @submit="onSubmit" :validation-schema="schema">
        <div class="form-row">
            <div class="form-group col">
                <label>Username</label>
                <Field name="username" type="text" class="form-control" />
            </div>
        </div>
        <div class="form-group">
            <button class="btn btn-primary" :disabled="isSubmitting">
                <span v-show="isSubmitting" class="spinner-border spinner-border-sm mr-1"></span>
                Save
            </button>
            <router-link to="/users" class="btn btn-link">Cancel</router-link>
        </div>
    </Form>
</template>
