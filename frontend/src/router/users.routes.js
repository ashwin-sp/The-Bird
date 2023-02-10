import { Layout, List, AddEdit, DeleteEdit} from '@/views/users';

export default {
    path: '/users',
    component: Layout,
    children: [
        { path: '', component: List },
        { path: 'add', component: AddEdit },
        { path: 'remove', component: DeleteEdit}
    ]
};
