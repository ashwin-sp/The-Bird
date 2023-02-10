<script setup>
const props = defineProps({
    PostID: {
        type: String,
        required: true
    },
    Username: {
        type: String,
        required: true
    },
    Timestamp: {
        type: String,
        required: true
    },
    Message: {
        type: String,
        required: true
    },
})
</script>

<template>
    <div class="card">
        <div class="card-header">
            <div class="container">
                <div class="col rows-col-1">
                    <div class="row">Username: {{ props.Username }}</div>
                </div>
            </div>
        </div>
        <div class="card-body">
            <h5 class="card-title">Message: {{ props.Message }}</h5>
        </div>
        <div class="card-footer text-muted">
            Timestamp: {{ props.Timestamp }}
        </div>
    </div>
</template>

<style scoped>
</style>
