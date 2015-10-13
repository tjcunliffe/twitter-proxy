import React from 'react'
import { Button, ButtonToolbar, Tooltip, OverlayTrigger } from 'react-bootstrap'
import ReactDOM from 'react-dom'


const updateComponent = function (that, data) {
    if (that.isMounted()) {

        let style = "";
        let caption = "";
        let tooltip = "";

        if (data.record) {
            // recording
            style = "warning";
            caption = "Recording";
            tooltip = "Proxy is currently in record mode. Press this button to start playback."

        } else {
            // playback
            style = "success";
            caption = "Playback";
            tooltip = "Proxy is currently in playback mode. Press this button to start recording."
        }

        that.setState({
            record: data.record,
            disabled: false,
            style: style,
            caption: caption,
            tooltipMessage: tooltip
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
            url: "/admin/state",
            tooltipMessage: ""
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
        var body = {
            record: !this.state.record
        };
        let that = this;
        $.ajax({
            type: "POST",
            dataType: "json",
            url: this.state.url,
            data: JSON.stringify(body),
            success: function (data) {
                updateComponent(that, data)
            }
        });

    },

    render() {
        const tooltip = (
            <Tooltip>{this.state.tooltipMessage}</Tooltip>
        );

        return (
            <OverlayTrigger placement='right' overlay={tooltip}>
                <Button onClick={this.handleClick} bsStyle={this.state.style}> {this.state.caption} </Button>
            </OverlayTrigger>
        )
    }

});

ReactDOM.render(<RecordToolbarComponent />, document.getElementById("app"));