// {people:[{id:number, name:string, imgs:string[]}], rankings: }


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


async function PopulateLeaderboard() {
    let { people, categories, elos } = await PostJSON("/api/leaderboard", {});
    for(ranking of elos) {
        let {id, elos} = ranking;
        
    }
}

window.onload = ()=> {
    PopulateLeaderboard();
}