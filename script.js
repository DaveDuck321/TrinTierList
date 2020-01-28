//Returns: {category:{id:int, name:string}, person1:{id:number, name:string, imgs:string[]}, ...}

let CurrentRank = {
    id: [],
    category: 0,
}

function Sample(list) {
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

async function ShowPeople(category) {
    document.getElementById("category").innerHTML = "Loading...";

    const data = await PostJSON("/people", {
        category: category
    });

    document.getElementById("category").innerHTML = data.category.name;

    const image_1 = document.querySelector("#first  img");
    const image_2 = document.querySelector("#second img");

    image_1.src = Sample(data.person1.imgs);
    image_2.src = Sample(data.person2.imgs);

    document.querySelector("#first  h3").innerText = data.person1.nickname;
    document.querySelector("#second h3").innerText = data.person2.nickname;

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

    document.querySelector("#first  button").onclick = () => { Vote(CurrentRank.id[0], CurrentRank.id[1], CurrentRank.category) };
    document.querySelector("#second button").onclick = () => { Vote(CurrentRank.id[1], CurrentRank.id[0], CurrentRank.category) };
};
