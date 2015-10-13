import React from 'react'
import { Button, ButtonToolbar } from 'react-bootstrap'
import ReactDOM from 'react-dom'


const updateComponent = function(that, data) {
    if (that.isMounted()) {

        let style = "";
        let caption = "";

        if(data.record) {
            // recording
            style = "warning";
            caption = "Recording"
        } else {
            // playback
            style = "success";
            caption = "Playback"
        }

        that.setState({
            record: data.record,
            disabled: false,
            style: style,
            caption: caption
        });
    }
};

const RecordToolbarComponent = React.createClass({
    getInitialState() {
        return {
            disabled: false,
            record: true,
            style: "warning",
            caption: "recording",
            url: "/admin/state"
        }
    },

    componentDidMount() {
        let that = this;
        $.ajax({
            type: "GET",
            dataType: "json",
            url: this.state.url,
            success: function (data) {
                updateComponent(that, data)
            }
        });
    },

    handleClick() {
        console.log(clicked)

    },

    render() {
        return <Button onClick={this.handleClick} bsStyle={this.state.style}> {this.state.caption} </Button>
    }

});

ReactDOM.render(<RecordToolbarComponent />, document.getElementById("app"));