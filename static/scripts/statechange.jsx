import React from 'react'
import { Button, ButtonToolbar } from 'react-bootstrap'

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
                console.log(data);

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
            }
        });

    },

    render() {
        return <Button bsStyle={this.state.style}> {this.state.caption} </Button>
    }

});

React.render(<RecordToolbarComponent />, document.getElementById("app"));