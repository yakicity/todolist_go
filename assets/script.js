const confirm_delete = (id) => {
    if(window.confirm(`Task ${id} を削除します．よろしいですか？`)) {
        location.href = `/task/delete/${id}`;
    }
}
 
const confirm_update = (id) => {
    if (window.confirm(`Task ${id} を変更します．よろしいですか？`)) {
        return true;
    }
    return false;
}

const confirm_userdelete = () => {
    if(window.confirm(`あなたのアカウントを削除します．よろしいですか？`)) {
        location.href = `/user/delete`;
    }
}

