//Returns: {category:{id:int, name:string}, person1:{id:number, name:string, imgs:string[]}, ...}

let imageCache = {};

let CurrentRank = {
    id: [],
    category: 0,
}

function randomItem(list) {
    return list[Math.floor(Math.random() * list.length)];
}

async function PostJSON(url, data) {
    const result = await fetch(url, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json;'
        },
        body: JSON.stringify(data),
    });
    const response = await result.json();
    if (!response.success) {
        console.error(response.msg);
    }
    return response;
}

function LoadImage(c2d, src) {
    c2d.src = src;
}

async function ShowPeople(category) {
    document.getElementById("category").innerHTML = "Loading...";

    const data = await PostJSON("/people", {
        category: category
    });

    document.getElementById("category").innerHTML = data.category.name;

    // TODO: Rename
    const c2d_1 = document.getElementById("person1-canvas");
    const c2d_2 = document.getElementById("person2-canvas");

    LoadImage(c2d_1, randomItem(data.person1.imgs));
    LoadImage(c2d_2, randomItem(data.person2.imgs));

    document.getElementById("person1-name").innerHTML = data.person1.nickname;
    document.getElementById("person2-name").innerHTML = data.person2.nickname;

    CurrentRank = {
        id: [data.person1.id, data.person2.id],
        category: data.category.id,
    };
}

async function Vote(won, lost, category) {
    PostJSON("/vote", {
        type: 'vote',

        won: won,
        lost: lost,

        category: category,
    });
    ShowPeople("random");
}

window.onload = () => {
    ShowPeople("random");

    document.getElementById("vote1").onclick = () => { Vote(CurrentRank.id[0], CurrentRank.id[1], CurrentRank.category) };
    document.getElementById("vote2").onclick = () => { Vote(CurrentRank.id[1], CurrentRank.id[0], CurrentRank.category) };
};
